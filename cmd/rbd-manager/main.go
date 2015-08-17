/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gambol99/rbd-fence/pkg/aws"
	"github.com/gambol99/rbd-fence/pkg/rbd"

	"github.com/golang/glog"
	"github.com/mitchellh/goamz/ec2"
)

var (
	// the rbd interface
	rbdClient rbd.RBDInterface
	// the aws events interface
	eventsClient aws.EC2EventsInterface
	// the hosts map
	hosts map[string]string
)

func main() {
	var err error
	flag.Parse()

	glog.Infof("Starting the %s Service, version: %s, git+sha: %s", Prog, Version, GitSha)

	// step: create the channel to termination requests
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if config.vpcID == "" {
		fmt.Printf("[error] you need to specify the vpc id for the instances are interested in")
		os.Exit(1)
	}

	// step: create a interface for events
	eventsClient, err = aws.NewEC2EventsInterface(config.aws_api_key, config.aws_api_secret,
		config.aws_region, config.vpcID)
	if err != nil {
		glog.Errorf("Failed to start service, error: %s", err)
		os.Exit(1)
	}

	// step: create the rbd interface
	rbdClient, err = rbd.NewRBDInterface()
	if err != nil {
		glog.Errorf("Failed to create interface to rbd command set, error: %s", err)
		os.Exit(1)
	}

	// step: create the event channels
	terminatedCh := eventsClient.AddEventListener(aws.STATUS_TERMINATED)
	stoppedCh := eventsClient.AddEventListener(aws.STATUS_STOPPED)
	runningCh := eventsClient.AddEventListener(aws.STATUS_RUNNING)
	// step: get a list of running hosts and their ip addresses
	hosts = eventsClient.GetRunningHosts()

	// step: enter the event loop: we are either listening to a termination signal, a terminated box
	// or a box being added
	for {
		select {
		// we have received a kill service signal
		case <-signalChannel:
			glog.Infof("Recieved a shutdown signal, exiting service")
			os.Exit(0)

		case event := <-runningCh:
			ev := (*aws.InstanceEvent)(event)
			glog.Infof("Region has a new instance running, instanceId: %s", ev.InstanceID)
			// add to the hosts map
			hosts[ev.Instance.InstanceId] = ev.Instance.PrivateIpAddress

		case event := <-stoppedCh:
			ev := (*aws.InstanceEvent)(event)
			glog.Infof("Region instance stopped, instanceId: %s", ev.InstanceID)
			go removeRBDLocks(&ev.Instance)

		case event := <-terminatedCh:
			ev := (*aws.InstanceEvent)(event)
			glog.Infof("Region instance terminated, instanceId: %s", ev.InstanceID)
			go removeRBDLocks(&ev.Instance)
		}
	}
}

// Checks to see if the instance has any locks and if so attempts to remove them
func removeRBDLocks(instance *ec2.Instance) {
	address, found := hosts[instance.InstanceId]
	if !found {
		glog.Errorf("The instance: %s was not found in the hosts map", instance.InstanceId)
		return
	}
	glog.Infof("Instance: %s, address: %s, state: %s, checking for locks", instance.InstanceId, address, instance.State.Name)

	var deleted = false

	for i := 0; i < 3; i++ {
		err := rbdClient.UnlockClient(address)
		if err != nil {
			glog.Errorf("Failed to unlock the images, attempting again if possible")
			<-time.After(time.Duration(5) * time.Second)
		}
		deleted = true
	}

	if !deleted {
		glog.Errorf("Failed to unlock any images that could have been held by client: %s", address)
	}

	// step: delete from the hosts map
	delete(hosts, instance.InstanceId)
}
