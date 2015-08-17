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
	"time"
)

var config struct {
	// the location / path of the rbd command
	rbd_path string
	// the aws key - should never really be used
	aws_api_key string
	// the aws secret
	aws_api_secret string
	// the api region
	aws_region string
	// the rbd poll
	rbd_pool string
	// the vpc id
	envTag string
}

const (
	DEFAULT_INTERVAL = time.Duration(1) * time.Minute
	DEFAULT_REGION   = "eu-west-1"
)

func init() {
	flag.StringVar(&config.aws_api_key, "key", "", "the aws api key to use (note: taken from env or iam is left empty)")
	flag.StringVar(&config.aws_api_secret, "secret", "", "the aws api secret, (note: taken from env or iam is left empty)")
	flag.StringVar(&config.aws_region, "region", DEFAULT_REGION, "the aws region we are speaking to")
	flag.StringVar(&config.rbd_pool, "pool", "rbd", "the pool the images live, leave black to check all pools")
	flag.StringVar(&config.envTag, "env", "", "the environment tag to filter out the instances, note any instance not tagged are ignored")
}
