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

package aws

import (
	"flag"
	"fmt"
	"time"
	"strings"

	gocache "github.com/pmylund/go-cache"
	"github.com/mitchellh/goamz/ec2"
	"github.com/golang/glog"
	"math/rand"
)

const (
	DEFAULT_POLL_INTERVAL = time.Duration(20) * time.Second
)

var ec2_instances_config struct {
	// the interval we should point the instance
	polling_interval time.Duration
}

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.DurationVar(&ec2_instances_config.polling_interval, "interval", DEFAULT_POLL_INTERVAL, "the default interval for polling instances")
}

// the implementation of a EC2InstancesInterface
type ec2Instances struct {
	// our interface to the api
	client EC2Interface
	// a in-memory cache for termination instances
	cache *gocache.Cache
	// a list of listeners for running instances
	listeners map[EventCh]int
	// a map of instances we have the details on
	hosts map[string]string
}

// Creates a new EC2InstanceInterface to consume events from
func NewEC2EventsInterface(aws_key, aws_secret, aws_region string) (EC2EventsInterface, error) {
	var err error
	glog.Infof("Creating a new EC2 Instances Interface for events")

	// step: create a new api for the service
	service := new(ec2Instances)
	service.listeners = make(map[EventCh]int, 0)
	service.hosts = make(map[string]string, 0)
	service.client, err = NewEC2Interface(aws_key, aws_secret, aws_region)
	if err != nil {
		return nil, err
	}
	service.cache = gocache.New(1 * time.Hour, 5 * time.Minute)

	// step: attempt to grab an initial state of running instances
	err = service.bootstrapRunningInstances(3)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap service, unable to retrieve runnings instance")
	}
	// step: start the synchronizing loop
	go service.synchronize()

	return service, nil
}

// synchronize ... the is main event loop, responsible for polling the ec2 api and looking for new
// and termination
func (r *ec2Instances) synchronize() {
	for {
		hosts_ids := make(map[string]*ec2.Instance, 0)

		// step: grab all the statuses of the instances
		running_now, err := r.client.DescribeAll()
		if err != nil {
			glog.Errorf("Failed to retrieve an updated list running instances, error: %s", err)
			// choice: we will continue and get them on the next run
			goto NEXT_LOOP
		}
		glog.V(5).Infof("Found %d instances presently in the region", len(running_now))

		// step: construct a map of the hosts for quick reference
		for _, x := range running_now {
			hosts_ids[x.InstanceId] = &x
		}

		// step: remove any instances which are no longer running ... i.e they were probably in a terminated state and
		// now aws has remove them
		for id, _ := range r.hosts {
			// step: is the hosts still in the list of instances?
			if _, found := hosts_ids[id]; found {
				continue
			}
			// step: the instance seems to have been remove, lets ensure it was terminated beforehand
			instance, found := r.getStatus(id)
			if !found {
				// choice: to dies as something very wrong has happened here
				glog.Fatalf("The instance: %s does not appear to be in the cache, killing myself", id)
			}

			// step: check the instance was in a termination state before
			if instance.State.Name != "terminated" {
				glog.Errorf("The instance: %s was not in a terminated state before, something strange is a foot", instance.InstanceId)
			} else {
				glog.Infof("The instance: %s has finally been removed from the terminated list", instance.InstanceId)
			}
			// step: delete from the cache and hosts map
			r.deleteStatus(id)
		}

		// step: we iterate the instances, checking for status changes or new instances
		for _, x := range running_now {
			// step: do we have a status for this instance in the cache or is it new?
			instance, found := r.getStatus(x.InstanceId)

			// the instance was not in the previous list - assuming it's new
			if !found {
				glog.V(2).Infof("Found a new instance: %s in the region, current status: %s", x.InstanceId, x.State.Name)
				// step: send the event
				r.sendEvent(&x, nil)
				// step: add the instance to the cache
				r.setStatus(x)
				continue
			}

			// has the state of the instance changed from before?
			if x.State.Name != instance.State.Name {
				glog.V(3).Infof("Status change for instance: %s, from: %s to: %s", instance.InstanceId, instance.State.Name, x.State.Name)
				// step: forward on the event
				r.sendEvent(&instance, &x)
				// step: update the instance in the cache
				r.setStatus(x)
				continue
			}

			// else the state of the instance has not changed
			glog.V(5).Infof("The instance: %s status has not changed, status: %s", x.InstanceId, x.State.Name)
		}

		NEXT_LOOP:
		// wait until the next polling
		<- time.After(ec2_instances_config.polling_interval)
	}
}

// AddEventListener ... add a event listener to the list of consumers
// 	filter:		an integer containing the bitwise filter
func (r *ec2Instances) AddEventListener(filter int) EventCh {
	// step: make a buffered channel for them
	ch := make(EventCh, 10)
	glog.V(4).Infof("Adding a event listner, channel: %v, filter: %d, filters: %s", ch, filter, r.filterToString(filter))
	r.listeners[ch] = filter
	return ch
}

// GetRunningHosts ... returns a list of running hosts
func (r *ec2Instances) GetRunningHosts() map[string]string {
	list := make(map[string]string, 0)
	for id, _ := range r.hosts {
		if instance, found := r.getStatus(id); found {
			if instance.State.Name == "running" {
				list[instance.InstanceId] = instance.PrivateIpAddress
			}
		}
	}
	return list
}

// Attempt to retrieve a list of running instances for bootstrapping purposes
func (r *ec2Instances) bootstrapRunningInstances(max_attempts int) error {
	for i := 0; i < max_attempts; i++ {
		glog.V(4).Infof("Attempting to retrieve the current instances from api, attmept: %d", i)
		instances, err := r.client.DescribeAll()
		if err != nil {
			glog.Errorf("Failed to retrieve instance details, error: %s", err)
			<- time.After(time.Duration(2) * time.Second)
			continue
		}
		// step: inject the instance statuses into the cache
		for _, x := range instances {
			r.setStatus(x)
		}
		return nil
	}
	return fmt.Errorf("failed to retrieve running instances from ec2")
}

// setStatus ... add the instance and status to the cache
func (r *ec2Instances) setStatus(instance ec2.Instance) {
	cache_key := r.getStatusKey(instance.InstanceId)
	glog.V(5).Infof("Adding instance status for instance: %s, status: %s, key: %s", instance.InstanceId,
		instance.State.Name, cache_key)
	r.cache.Set(r.getStatusKey(instance.InstanceId), instance, gocache.NoExpiration)
	// step: update the map of instance we have
	r.hosts[instance.InstanceId] = cache_key
}

// getStatus ... retrieve the statue from the cache
func (r *ec2Instances) getStatus(id string) (ec2.Instance, bool) {
	cache_key := r.getStatusKey(id)
	glog.V(5).Infof("Looking for the status of instance: %s, key: %s", id, cache_key)
	if x, found := r.cache.Get(cache_key); found {
		instance := x.(ec2.Instance)
		return instance, true
	}
	return ec2.Instance{}, false
}

// deleteStatus ... remove the status from the cache
func (r *ec2Instances) deleteStatus(id string) {
	cache_key := r.getStatusKey(id)
	glog.V(4).Infof("Deleting the status on the instance: %s, key: %s", id, cache_key)
	// step: delete from cache
	r.cache.Delete(cache_key)
	delete(r.hosts, id)
}

// getStatusKey ... construct a status key from the instance
func (r ec2Instances) getStatusKey(id string) string {
	return fmt.Sprintf("status_%s", id)
}

// sendEvent ... iterated the listener, finds those whom match the filter and sends the event
func (r ec2Instances) sendEvent(from, to *ec2.Instance) {
	var state int
	// step: if no to instance, it's because it's a new instance
	if to == nil {
		state = r.convertStatusToFilter(from.State.Name)
		to = from
	} else {
		state = r.convertStatusToFilter(to.State.Name)
	}
	// step: construct the event
	event := &InstanceEvent{
		InstanceID: from.InstanceId,
		EventType: state,
		Instance: *to,
	}
	// step: iterate the listeners and find consumer interested in this event
	for ch, filter := range r.listeners {
		glog.V(4).Infof("Checking filter: %d (%s) <-> state: %d (%s|%s)", filter, r.filterToString(filter),
			state, r.filterToString(state), to.State.Name)
		if state&filter != 0 {
			glog.V(4).Infof("Found match for state: %d, filter: %d, channel: %v", state, filter, ch)
			// step: send the event upstream
			go func(ln EventCh) {
				glog.V(5).Infof("Sending the event: %s to consumer: %v", event, ln)
				ln <- event
			}(ch)
		}
	}
}

func (r ec2Instances) convertStatusToFilter(status string) int {
	if status == "running" {
		return STATUS_RUNNING
	}
	if status == "stopped" {
		return STATUS_STOPPED
	}
	if status == "terminated" {
		return STATUS_TERMINATED
	}
	if status == "stopping" {
		return STATUS_STOPPING
	}
	if status == "pending" {
		return STATUS_PENDING
	}
	if status == "shutting-down" {
		return STATUS_SHUTTING_DOWN
	}
	return STATUS_UNKNOWN
}

func (r ec2Instances) filterToString(filter int) string {
	var filters []string
	if (filter&STATUS_RUNNING) == STATUS_RUNNING {
		filters = append(filters, "running")
	}
	if (filter&STATUS_STOPPED) == STATUS_STOPPED {
		filters = append(filters, "stopped")
	}
	if (filter&STATUS_SHUTTING_DOWN) == STATUS_SHUTTING_DOWN {
		filters = append(filters, "shutting-down")
	}

	if (filter&STATUS_STOPPING) == STATUS_STOPPING {
		filters = append(filters, "stopping")
	}
	if (filter&STATUS_TERMINATED) == STATUS_TERMINATED {
		filters = append(filters, "terminated")
	}
	if (filter&STATUS_PENDING) == STATUS_PENDING {
		filters = append(filters, "pending")
	}
	if (filter&STATUS_UNKNOWN) == STATUS_UNKNOWN {
		filters = append(filters, "unknown")
	}
	return strings.Join(filters, ",")
}
