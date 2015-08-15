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
	"github.com/golang/glog"
	"github.com/mitchellh/goamz/aws"
)

const (
	FILTER_RUNNING 	  = "running"
	FILTER_TERMINATED = "terminated"
)

type ec2Helper struct {
	// the client for ec2
	client *ec2.EC2
	// the region
	region string
}

// Creates a new EC2 Helper interface
func NewEC2Interface(aws_key, aws_secret, region string) (EC2Interface, error) {
	glog.Infof("Create a new EC2 API client for region: %s", region)

	service := new(ec2Helper)
	service.region = region
	// step: check the region is valid
	aws_region, valid := service.isValidRegion(region)
	if !valid {
		return nil, fmt.Errorf("invalid aws region specified, please check")
	}

	// step: attempt to acquire credentials
	auth, err := aws.GetAuth(aws_key, aws_secret)
	if err != nil {
		return nil, fmt.Errorf("unable to find authentication details, error: %s", err)
	}

	// step: create a client and return
	service.client = ec2.New(auth, aws_region)
	glog.V(5).Infof("Successfully create a api client for aws in region: %s", aws_region)

	return service, nil
}

// Get a complete list of instances
func (r ec2Helper) DescribeInstances(filter *ec2.Filter) ([]ec2.Instance, error) {
	var hosts = make([]ec2.Instance, 0)
	glog.V(5).Infof("Retreiving a list of instances from EC2, instance filter: %V", filter)

	// step: call the api and retrieve the results
	result, err := r.client.Instances([]string{}, filter)
	if err != nil {
		return hosts, err
	}
	glog.V(5).Infof("Found %d instances matching the filter %V", len(result.Reservations), filter)

	// step: if we don't have any running instance, return empty
	if len(result.Reservations) <= 0 || len(result.Reservations[0].Instances) <= 0 {
		return hosts, nil
	}

	// step: extract and add hosts. Almost certainly a better way of doing this
	for _, instances := range result.Reservations {
		for _, x := range instances.Instances {
			hosts = append(hosts, x)
		}
	}

	return hosts, nil
}

// Terminate the instance
func (r ec2Helper) TerminatedInstance(id string) error {
	glog.Infof("Terminating the instance: %s in region: %s", id, r.region)
	// step: check the instance exists first
	if found, err := r.Exists(id); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("the instance: %s does not exist in the resgion", id)
	}
	// step: terminate the instance
	_, err := r.client.TerminateInstances([]string{id})
	if err != nil {
		return err
	}
	return nil
}

// Check an instances exists in the region
func (r ec2Helper) Exists(id string) (bool, error) {
	// step: get a list of all instances
	instances, err := r.DescribeAll()
	if err != nil {
		return false, err
	}
	glog.V(5).Infof("Checking if the instance: %s exists in the region", id)
	for _, x := range instances {
		if x.InstanceId == id {
			return true, nil
		}
	}
	return false, nil
}

// Get all instances
func (r ec2Helper) DescribeAll() ([]ec2.Instance, error) {
	return r.DescribeInstances(ec2.NewFilter())
}

// Get running instances
func (r ec2Helper) DescribeRunning() ([]ec2.Instance, error) {
	filter := ec2.NewFilter()
	filter.Add("instance-state-name", FILTER_RUNNING)
	return r.DescribeInstances(filter)
}

// Get terminated instances
func (r ec2Helper) DescribeTerminated() ([]ec2.Instance, error) {
	filter := ec2.NewFilter()
	filter.Add("instance-state-name", FILTER_TERMINATED)
	return r.DescribeInstances(filter)
}

func (r ec2Helper) isValidRegion(region string) (aws.Region, bool) {
	x, found := aws.Regions[region]
	return x, found
}
