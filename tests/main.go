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
	"strings"
	"fmt"
	"os/exec"
	"os"
)

// event types
const (
	STATUS_RUNNING    = 1 << iota
	STATUS_STOPPING
	STATUS_STOPPED
	STATUS_TERMINATED
	STATUS_PENDING
	STATUS_UNKNOWN
)

func main() {
	isFilter(STATUS_RUNNING, STATUS_RUNNING|STATUS_PENDING)
	isFilter(STATUS_RUNNING, STATUS_TERMINATED)
	isFilter(STATUS_TERMINATED, STATUS_RUNNING)
	isFilter(STATUS_STOPPING, STATUS_RUNNING)
	isFilter(STATUS_TERMINATED,STATUS_TERMINATED|STATUS_RUNNING|STATUS_PENDING)

	out, err := exec.Command("ls", "-l", "-s").Output()
	if err != nil {
		fmt.Printf("failed to execute command, error: %s", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func isFilter(state, filter int) bool {
	fmt.Printf("state: %d (%s) <-> filter: %d (%s), ", state, filterToString(state), filter, filterToString(filter))
	if state&filter != 0 {
		fmt.Println("MATCHED")
		return true
	}
	fmt.Println("NO MATCH")
	return false
}

func convertStatusToFilter(status string) int {
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
	return STATUS_UNKNOWN
}

func filterToString(filter int) string {
	var filters []string
	if (filter&STATUS_RUNNING) == STATUS_RUNNING {
		filters = append(filters, "running")
	}
	if (filter&STATUS_STOPPED) == STATUS_STOPPED {
		filters = append(filters, "stopped")
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
