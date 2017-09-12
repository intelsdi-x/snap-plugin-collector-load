// +build linux freebsd

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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
	"os"

	"github.com/intelsdi-x/snap/control/plugin"

	"github.com/intelsdi-x/snap-plugin-collector-load/load"
)

func main() {
	loadPlugin := load.New()
	if loadPlugin == nil {
		panic("Failed to initialize plugin!\n")
	}

	plugin.Start(
		load.Meta(),
		loadPlugin,
		os.Args[1],
	)
}
