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
)

const (
	CEPH_POOLS_OUTPUT = `epoch 62
	fsid 1b93e22a-14ea-4874-a149-0bcd5cc7ad18
created 2015-08-12 11:55:18.108291
modified 2015-08-14 09:38:23.439233
flags
pool 0 'rbd' replicated size 2 min_size 1 crush_ruleset 0 object_hash rjenkins pg_num 64 pgp_num 64 last_change 1 flags hashpspool stripe_width 0
pool 1 'data' replicated size 2 min_size 1 crush_ruleset 0 object_hash rjenkins pg_num 128 pgp_num 128 last_change 24 flags hashpspool crash_replay_interval 45 stripe_width 0
pool 2 'metadata' replicated size 2 min_size 1 crush_ruleset 0 object_hash rjenkins pg_num 128 pgp_num 128 last_change 23 flags hashpspool stripe_width 0
max_osd 3
osd.0 up   in  weight 1 up_from 10 up_thru 29 down_at 0 last_clean_interval [0,0) 10.50.21.100:6800/1 10.50.21.100:6801/1 10.50.21.100:6802/1 10.50.21.100:6803/1 exists,up 83ffbeda-21ba-41e4-8171-e95bb213d6d9
osd.1 up   in  weight 1 up_from 14 up_thru 29 down_at 0 last_clean_interval [0,0) 10.50.20.100:6800/1 10.50.20.100:6801/1 10.50.20.100:6802/1 10.50.20.100:6803/1 exists,up 605915b5-926b-41d7-95d0-8cd07a9ac8f1
osd.2 up   in  weight 1 up_from 18 up_thru 29 down_at 0 last_clean_interval [0,0) 10.50.22.100:6800/1 10.50.22.100:6801/1 10.50.22.100:6802/1 10.50.22.100:6803/1 exists,up a4ab6e70-5eba-42a1-b049-0549b9c32012
`
)
