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
	"flag"
	"os"

	"github.com/gambol99/rbd-fence/pkg/rbd"

	"github.com/golang/glog"
)

var config struct {
	// the ip address of the client
	address string
}

func init() {
	flag.StringVar(&config.address, "ip", "", "the ip address of the client which you wish to unlock")
}

func main() {
	flag.Parse()
	if config.address == "" {
		glog.Errorf("You have not specified the ip address of the client to remove lock ownership")
		os.Exit(1)
	}

	// step: grab a rbd interface
	client, err := rbd.NewRBDInterface()
	if err != nil {
		glog.Errorf("Failed to create a client interface for rbd, error: %s", err)
		os.Exit(1)
	}

	err = client.UnlockClient(config.address)
	if err != nil {
		glog.Errorf("Failed to unlock the images held by %s", config.address)
		os.Exit(1)
	}

	glog.Infof("Successfully remove any locks")
}
