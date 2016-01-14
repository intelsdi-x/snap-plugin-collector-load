// +build linux

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

package load

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	str "github.com/intelsdi-x/snap-plugin-utilities/strings"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	// VENDOR namespace part
	VENDOR = "intel"
	// OS namespace part
	OS = "linux"
	// PLUGIN namespace part
	PLUGIN = "load"
	// VERSION of load info plugin
	VERSION = 1
)

var loadInfo = "/proc/loadavg"

type loadPlugin struct {
	stats map[string]interface{}
	host  string
	cpus  int
}

// New create instance of load info plugin
func New() *loadPlugin {
	fh, err := os.Open(loadInfo)

	if err != nil {
		return nil
	}
	defer fh.Close()

	cpu, err := getCPUs()

	if err != nil {
		return nil
	}

	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}

	mp := &loadPlugin{stats: map[string]interface{}{}, host: host, cpus: cpu}

	return mp
}

func getCPUs() (int, error) {
	out, err := exec.Command("lscpu", "-p").Output()
	if err != nil {
		return -1, err
	}

	lines := strings.Split(string(out), "\n")
	lines = str.Filter(lines, func(s string) bool {
		return s != ""
	})
	last := lines[len(lines)-1]
	cpus, err := strconv.Atoi(strings.Split(last, ",")[0])

	if err != nil {
		return -1, err
	}

	return cpus + 1, nil
}

func getStats(stats map[string]interface{}, cpus int) error {
	fh, err := os.Open(loadInfo)

	if err != nil {
		return err
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 5 {
			return fmt.Errorf("Wrong %s format", loadInfo)
		}

		min1, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return err
		}
		stats["min1"] = min1
		stats["min1_rel"] = min1 / float64(cpus)

		min5, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return err
		}
		stats["min5"] = min5
		stats["min5_rel"] = min5 / float64(cpus)

		min15, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return err
		}
		stats["min15"] = min15
		stats["min15_rel"] = min15 / float64(cpus)

		stats["scheduling"] = fields[3]
	}

	return nil
}

// GetMetricTypes returns list of available metric types
// It returns error in case retrieval was not successful
func (mp *loadPlugin) GetMetricTypes(_ plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	metricTypes := []plugin.PluginMetricType{}
	if err := getStats(mp.stats, mp.cpus); err != nil {
		return nil, err
	}
	for stat := range mp.stats {
		metricType := plugin.PluginMetricType{Namespace_: []string{VENDOR, OS, PLUGIN, stat}}
		metricTypes = append(metricTypes, metricType)
	}
	return metricTypes, nil
}

// CollectMetrics returns list of requested metric values
// It returns error in case retrieval was not successful
func (mp *loadPlugin) CollectMetrics(metricTypes []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := []plugin.PluginMetricType{}
	getStats(mp.stats, mp.cpus)
	for _, metricType := range metricTypes {
		ns := metricType.Namespace()
		if len(ns) < 4 {
			return nil, fmt.Errorf("Namespace length is too short (len = %d)", len(ns))
		}
		stat := ns[3]
		val, ok := mp.stats[stat]
		if !ok {
			return nil, fmt.Errorf("Requested stat %s is not available!", stat)
		}
		metric := plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      val,
			Source_:    mp.host,
			Timestamp_: time.Now(),
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

// GetConfigPolicy returns config policy
// It returns error in case retrieval was not successful
func (mp *loadPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}
