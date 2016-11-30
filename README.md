# Snap collector plugin - load
This plugin collects metrics from /proc/loadavg kernel interface about load average figures giving the number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 1, 5, and 15 minutes.  

It's used in the [Snap framework](http://github.com:intelsdi-x/snap).

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Operating systems](#operating-systems)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license-and-authors)
6. [Acknowledgements](#acknowledgements)

## Getting Started
### System Requirements
* [golang 1.6+](https://golang.org/dl/)

### Operating systems
All OSs currently supported by Snap:
* Linux/amd64

### Installation
#### Download the plugin binary:

You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-collector-load/releases) page. Download the plugin from the latest release and load it into `snapteld` (`/opt/snap/plugins` is the default location for Snap packages).

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-load

Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-load.git
```

Build the Snap load plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `./build/`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap#getting-started)
* Load the plugin and create a task, see example in [Examples](#examples).

Plugin may be additionally configured with path to `/proc/loadavg` file. It may be useful for running plugin in Docker container.
Example configuration provided in [Examples](#examples)
In case configuration is not provided in task manifest, plugin will use default path `/proc/loadavg`

## Documentation

### Collected Metrics
This plugin has the ability to gather the following metrics:

Namespace | Description (optional)
----------|-----------------------
/intel/procfs/load/min1 | number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 1 minute
/intel/procfs/load/min5 | number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 5 minutes
/intel/procfs/load/min15 | number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 15 minutes
/intel/procfs/load/min1_rel | number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 1 minute per core
/intel/procfs/load/min5_rel | number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 5 minutes per core
/intel/procfs/load/min15_rel | number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 15 minutes per core
/intel/procfs/load/runnable_scheduling | The number of currently runnable kernel scheduling entities (processes, threads).
/intel/procfs/load/existing_scheduling | The number of kernel scheduling entities that currently exist on the system

### Examples
Example of running Snap load collector and writing data to file.

In terminal go to [examples/tasks](examples/tasks) directory and execute `run-mock-load.sh`:
```
$ cd examples/tasks && ./run-mock-load.sh
```

Script will start docker container in which it will download load collector, passthru processor and file publisher.
Then it will start Snap daemon, load all plugins and start example [task](examples/tasks/task-load.json).

Type `snaptel task list` to get the list of active tasks
```
$ snaptel task list
ID                                       NAME                                            STATE           HIT     MISS    FAIL    CREATED                 LAST FAILURE
95b9fd8b-42d8-4836-be08-3865ba1f7926     Task-95b9fd8b-42d8-4836-be08-3865ba1f7926       Running         146     0       0       11:44AM 11-17-2016
```

See realtime output from `snaptel task watch <task_id>` (CTRL+C to exit)
```
$ snaptel task watch 95b9fd8b-42d8-4836-be08-3865ba1f7926
Watching Task (95b9fd8b-42d8-4836-be08-3865ba1f7926):
NAMESPACE                                DATA    TIMESTAMP
^Cntel/procfs/load/min1                  0.38    2016-11-17 11:47:45.801133834 +0000 UTC
/intel/procfs/load/min15                 0.43    2016-11-17 11:47:45.801146089 +0000 UTC
/intel/procfs/load/runnable_scheduling   1       2016-11-17 11:47:45.801163284 +0000 UTC
```

Stop task:
```
$ snaptel task stop 02dd7ff4-8106-47e9-8b86-70067cd0a850
Task stopped:
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
```

Type 'exit' when your done

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-load/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-load/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap.

To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements
* Author: [Marcin Krolik](https://github.com/marcin-krolik/)
