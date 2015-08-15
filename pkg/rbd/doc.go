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

import "fmt"

type CephPool struct {
	// the id of the pool
	PoolNum int `json:"poolnum"`
	// the name of the pool
	Name string `json:"poolname"`
}

type RbdImage struct {
	// the name of the image
	Name string `json:"image"`
	// the size
	Size int64 `json:"size"`
	// the format version
	Format int `json:"format"`
	// a exclusive lock
	LockType string `json:"lock_type"`
}

func (r RbdImage) IsLocked() bool {
	if r.LockType == "exclusive" {
		return true
	}
	return false
}

type RbdOwner struct {
	// the lockId on the device
	LockID string
	// the client id
	ClientID string
	// the address of the owner
	Address string
	// the session
	Session string
}

func (r RbdOwner) String() string {
	return fmt.Sprintf("lockID: '%s', clientID: '%s', address: '%s', session: '%s'",
		r.LockID, r.ClientID, r.Address, r.Session)
}

// the interface to RBD commands
type RBDInterface interface {
	// Get a list of the pool
	GetPools() ([]CephPool, error)
	// Get the owner of the lock
	GetLockOwner(RbdImage, CephPool) (RbdOwner, error)
	// Get a list of the images
	GetImages(CephPool) ([]RbdImage, error)
	// Unlock a image
	UnlockImage(RbdImage, CephPool) error
	// Unlock a client
	UnlockClient(string) error
}
