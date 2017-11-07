// +build linux

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

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/str"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"
)

const (
	// VENDOR namespace part
	pluginVendor = "intel"
	// FS namespace part
	fs = "procfs"
	// PLUGIN namespace part
	pluginName = "load"
	// VERSION of load info plugin
	pluginVersion = 3
)

var pluginInfoFields = map[string]infoFields{
	"min1": infoFields{
		description: "number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 1 minute",
		unit:        "",
	},
	"min5": infoFields{
		description: "number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 5 minutes",
		unit:        "",
	},
	"min15": infoFields{
		description: "number of jobs in the run queue (state R) or waiting for disk I/O (state D) averaged over 15 minutes",
		unit:        "",
	},
	"runnable_scheduling": infoFields{
		description: "The number of currently runnable kernel scheduling entities (processes, threads)",
		unit:        "",
	},
	"existing_scheduling": infoFields{
		description: "The number of kernel scheduling entities that currently exist on the system",
		unit:        "",
	},
}

// New create instance of load info plugin
func New() *loadPlugin {
	logger := log.New()
	cpu, err := getCPUs()
	if err != nil {
		logger.Errorf("Error while reading number of cpus {%s}", err)
		return nil
	}

	lp := &loadPlugin{cpus: cpu, logger: logger}
	lp.logger.Debug("New plugin instance created")
	return lp
}

// Meta returns plugin meta data
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		pluginName,
		pluginVersion,
		plugin.CollectorPluginType,
		[]string{},
		[]string{plugin.SnapGOBContentType},
		plugin.ConcurrencyCount(1),
	)
}

// GetMetricTypes returns list of available metric types
// It returns error in case retrieval was not successful
func (lp *loadPlugin) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	lp.logger.Debug("Calling GetMetricTypes()")
	stats := LoadMetrics{}
	namespaces := []string{}
	metricTypes := []plugin.MetricType{}

	ns.FromCompositionTags(stats, strings.Join([]string{pluginVendor, fs, pluginName}, "/"), &namespaces)
	lp.logger.Debugf("Namespaces created %v", namespaces)

	for _, namespace := range namespaces {
		last := path.Base(namespace)
		info := getInfoFields(last)
		metricType := plugin.MetricType{
			Namespace_:   core.NewNamespace(strings.Split(namespace, "/")...),
			Description_: info.description,
			Unit_:        info.unit,
			Config_:      cfg.ConfigDataNode,
		}
		metricTypes = append(metricTypes, metricType)
		lp.logger.Debugf("MetricType created successfully %s", metricType.Namespace().String())
	}

	return metricTypes, nil
}

// CollectMetrics returns list of requested metric values
// It returns error in case retrieval was not successful
func (lp *loadPlugin) CollectMetrics(metricTypes []plugin.MetricType) ([]plugin.MetricType, error) {
	lp.logger.Debug("Calling CollectMetrics()")
	metrics := []plugin.MetricType{}
	stats := LoadMetrics{}

	// get location of loadavg
	procPath, _ := config.GetConfigItem(metricTypes[0], "proc_path")
	lp.logger.Debugf("Procfs loadavg location found %s", procPath.(string))

	// read metrics from provided loadavg loacation
	if err := getStats(procPath.(string), &stats, lp.cpus); err != nil {
		lp.logger.Errorf("Could not read metrics from %s location. {%s}", procPath.(string), err)
		return nil, err
	}

	for _, metricType := range metricTypes {
		namespace := metricType.Namespace()
		if len(namespace.Strings()) < 4 {
			lp.logger.Errorf("Namespace length is too short (len = %d)", len(namespace.Strings()))
			return nil, fmt.Errorf("Namespace length is too short (len = %d)", len(namespace.Strings()))
		}

		val := ns.GetValueByNamespace(stats, namespace.Strings()[3:])
		lp.logger.Debugf("Found value %v for %s", val, namespace.String())
		metric := plugin.MetricType{
			Namespace_: namespace,
			Data_:      val,
			Timestamp_: time.Now(),
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

// GetConfigPolicy returns config policy
// It returns error in case retrieval was not successful
func (lp *loadPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	lp.logger.Debug("Creating new rule for proc_path")
	rule, _ := cpolicy.NewStringRule("proc_path", false, "/proc")
	node := cpolicy.NewPolicyNode()
	node.Add(rule)
	cp.Add([]string{pluginVendor, fs, pluginName}, node)
	return cp, nil
}

type loadPlugin struct {
	cpus   int
	logger *log.Logger
}

type infoFields struct {
	description string
	unit        string
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

func getStats(procPath string, stats *LoadMetrics, cpus int) error {
	content, err := ioutil.ReadFile(path.Join(procPath, "loadavg"))
	if err != nil {
		return err
	}

	fields := strings.Fields(string(content))

	if len(fields) < 5 {
		return fmt.Errorf("Wrong %s format", string(content))
	}

	stats.Min1, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return err
	}
	stats.Min1Rel = stats.Min1 / float64(cpus)

	stats.Min5, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return err
	}
	stats.Min5Rel = stats.Min5 / float64(cpus)

	stats.Min15, err = strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return err
	}
	stats.Min15Rel = stats.Min15 / float64(cpus)

	scheduling := strings.Split(fields[3], "/")
	if len(scheduling) != 2 {
		return fmt.Errorf("Scheduling data format incorrect {%s}", fields[3])
	}

	stats.RunSched, err = strconv.Atoi(scheduling[0])
	if err != nil {
		return err
	}

	stats.ExistingSched, err = strconv.Atoi(scheduling[1])
	if err != nil {
		return err
	}

	return nil
}

func getInfoFields(metric string) infoFields {
	info, ok := pluginInfoFields[metric]
	if !ok {
		info = infoFields{description: "", unit: ""}
	}
	return info
}
