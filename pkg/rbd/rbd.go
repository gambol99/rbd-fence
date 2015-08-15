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

package rbd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gambol99/rbd-fence/pkg/utils"

	"github.com/golang/glog"
)

type rbdUtil struct{}

var (
	lockRegex      = regexp.MustCompile("^(client\\.[0-9]+)\\s+([[:alnum:]\\._-]*)\\s+([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}):[0-9]+/([0-9]+)\\s+$")
	defaultTimeout = time.Duration(10) * time.Second
)

// NewRBDInterface ... create a new service interface for rbd operations
func NewRBDInterface() (RBDInterface, error) {
	return &rbdUtil{}, nil
}

// Get a list of the pool
func (r rbdUtil) GetPools() ([]CephPool, error) {
	// step: get the pool output
	result, err := utils.Execute(defaultTimeout, "ceph", "osd", "lspools", "-f", "json")
	if err != nil {
		return nil, err
	}

	// step: parse the json output
	var pools []CephPool
	err = json.NewDecoder(strings.NewReader(string(result))).Decode(&pools)
	if err != nil {
		return nil, err
	}
	return pools, nil
}

func (r rbdUtil) GetImages(pool CephPool) ([]RbdImage, error) {
	// step: get the pool output
	result, err := utils.Execute(defaultTimeout, "rbd", "-p", pool.Name, "ls", "-l", "--format", "json")
	if err != nil {
		return nil, err
	}

	// step: parse the json output
	var images []RbdImage
	err = json.NewDecoder(strings.NewReader(string(result))).Decode(&images)
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (r rbdUtil) GetLockOwner(image RbdImage, pool CephPool) (RbdOwner, error) {
	var owner RbdOwner

	// step: construct the command
	output, err := utils.Execute(defaultTimeout, "rbd", "-p", pool.Name, "lock", "list", image.Name)
	if err != nil {
		return owner, fmt.Errorf("%s, output: %s", err, output)
	}

	// step: parse the output
	reader := bufio.NewReader(strings.NewReader(string(output)))
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if matched := lockRegex.Match([]byte(line)); matched {
			matches := lockRegex.FindAllStringSubmatch(line, -1)
			owner.ClientID = matches[0][1]
			owner.LockID = matches[0][2]
			owner.Address = matches[0][3]
			owner.Session = matches[0][4]
		}
	}

	return owner, nil
}

// UnlockImage ... removes a rbd lock from the image
//	image:	the details of the image (name/pool) etc that you wish to remove the lock
func (r rbdUtil) UnlockImage(image RbdImage, cephPool CephPool) error {
	var name = image.Name
	var pool = cephPool.Name

	glog.Infof("Removing the lock on image: %s/%s", pool, name)
	// step: we get the owner of the image
	owner, err := r.GetLockOwner(image, cephPool)
	if err != nil {
		glog.V(4).Infof("Failed to get the lock owner of image: %s/%s, error: %s", pool, name, err)
		return err
	}

	// step: construct the command
	output, err := utils.Execute(defaultTimeout, "rbd", "-p", pool, "lock", "remove", name, owner.LockID, owner.ClientID)
	if err != nil {
		return fmt.Errorf("%s, output: %s", err, output)
	}

	return nil
}

// UnlockClient ... find any images which have been locked by the client ip address and removes them
func (r rbdUtil) UnlockClient(address string) error {
	glog.V(3).Infof("Attemping to remove any lock for client: %s", address)

	// step: this is horrid, but will suffice for now
	pools, err := r.GetPools()
	if err != nil {
		return err
	}

	// step: iterate the pool and looks for images
	for _, pool := range pools {

		// step: get all the images for this pool
		images, err := r.GetImages(pool)
		if err != nil {
			continue
		}

		// step: check if any of a lock and it any of them were locked by the instance address
		for _, image := range images {

			if !image.IsLocked() {
				glog.V(5).Infof("Skipping the image: %s/%s as it is not locked", pool.Name, image.Name)
				continue
			}

			// step: get the owner
			owner, err := r.GetLockOwner(image, pool)
			if err != nil {
				glog.Errorf("Failed to get the owner of the image: %s/%s, error: %s", pool.Name, image.Name, err)
				continue
			}
			// step: is the owner us?
			if owner.Address == address {
				glog.V(4).Infof("Client: %s has image: %s/%s locked, attempting to remove lock", address, pool.Name, image.Name)
				// we need to unlock the image
				err := r.UnlockImage(image, pool)
				if err != nil {
					glog.Errorf("Failed to unable the image: %s/%s, error: %s", pool.Name, image.Name, err)
					continue
				}
				glog.Infof("Successfully removed the lock on %s/%s from client: %s", pool.Name, image.Name, address)
			}
		}
	}

	return nil
}
