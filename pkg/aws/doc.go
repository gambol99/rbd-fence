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
	"fmt"

	"github.com/mitchellh/goamz/ec2"
)

// event types
const (
	STATUS_RUNNING = 1 << iota
	STATUS_STOPPING
	STATUS_STOPPED
	STATUS_SHUTTING_DOWN
	STATUS_TERMINATED
	STATUS_PENDING
	STATUS_UNKNOWN
)

// InstanceEvent ... the stucture for a instance event
type InstanceEvent struct {
	// the instance id
	InstanceID string
	// the event type
	EventType int
	// the instance if required
	Instance ec2.Instance
}

func (r InstanceEvent) String() string {
	return fmt.Sprintf("instanceId: %s, type: %d", r.InstanceID, r.EventType)
}

// EventCh ... a channel to receive events upon
type EventCh chan *InstanceEvent

// EC2Interface ... a helper interface to ec2 instances
type EC2Interface interface {
	// Get a complete list of instances
	DescribeInstances(*ec2.Filter) ([]ec2.Instance, error)
	// Get running instances
	DescribeRunning() ([]ec2.Instance, error)
	// Get terminated instances
	DescribeTerminated() ([]ec2.Instance, error)
	// Get all instances
	DescribeAll() ([]ec2.Instance, error)
	// terminate the instance
	TerminatedInstance(string) error
	// check the instance exists
	Exists(string) (bool, error)
}

// EC2EventsInterface ... the interface for a instance listener
type EC2EventsInterface interface {
	// Add a event listener for terminated instances
	AddEventListener(int) EventCh
	// Get running hosts
	GetRunningHosts() map[string]string
}
