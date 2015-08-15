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

package utils

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
)

// ExecCommand ... executes a command and returns the output
// 	command:	the command you wish to execute
//	args:		an array of argument to pass to the command
// 	timeout:	a time.Duration to wait before killing off the command
func Execute(timeout time.Duration, command string, args ...string) ([]byte, error) {
	var b bytes.Buffer

	if timeout.Seconds() <= 0 {
		timeout = time.Duration(10) * time.Second
	}

	glog.V(5).Infof("Attempting to execute the command: %s, args: [%s], timeout: %s", command,
		strings.Join(args, ","), timeout.String())

	cmd := exec.Command(command, args...)
	cmd.Stdout = &b
	cmd.Stderr = &b
	cmd.Start()
	timer := time.AfterFunc(timeout, func() {
		err := cmd.Process.Kill()
		if err != nil {
			panic(err)
		}
	})
	err := cmd.Wait()
	if err != nil {
		glog.Errorf("Failed to execute the command, error: %s", err)
		return []byte{}, err
	}
	// stop the timer
	timer.Stop()

	glog.V(5).Infof("Command output: %s", b.String())

	return b.Bytes(), nil
}
