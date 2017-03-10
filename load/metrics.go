// +build linux freebsd

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package load

// LoadMetrics stores load average figures of the system
type LoadMetrics struct {
	// Min1 is load average the number of jobs in the run queue or waiting for disk I/O averaged over 1 minute
	Min1 float64 `json:"min1"`
	// Min5 is load average the number of jobs in the run queue or waiting for disk I/O averaged over 5 minutes
	Min5 float64 `json:"min5"`
	// Min15 is load average the number of jobs in the run queue or waiting for disk I/O averaged over 15 minutes
	Min15 float64 `json:"min15"`
	// Min1 is load average the number of jobs in the run queue or waiting for disk I/O averaged over 1 minute per core
	Min1Rel float64 `json:"min1_rel"`
	// Min5 is load average the number of jobs in the run queue or waiting for disk I/O averaged over 5 minutes per core
	Min5Rel float64 `json:"min5_rel"`
	// Min15 is load average the number of jobs in the run queue or waiting for disk I/O averaged over 15 minutes per core
	Min15Rel float64 `json:"min15_rel"`
	// RunSched is the number of currently runnable kernel scheduling entities
	RunSched int `json:"runnable_scheduling"`
	// ExistingSched the number of kernel scheduling entities that currently exist on the system
	ExistingSched int `json:"existing_scheduling"`
}
