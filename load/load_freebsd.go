// +build freebsd

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation
Copyright 2017 Steven Wills

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

/*
#cgo LDFLAGS: -lkvm -lelf

#include <sys/param.h>
#include <sys/user.h>
#include <paths.h>
#include <fcntl.h>
#include <kvm.h>

#include <sys/cdefs.h>
#include <sys/types.h>
#include <sys/resource.h>
#include <sys/sysctl.h>
#include <vm/vm_param.h>
#include <limits.h>

static double avenrun[3];
static kvm_t *kd;
static int nproc = 0;
static int nthrd = 0;

int
mygetloadavg(double loadavg[])
{
        struct loadavg loadinfo;
        int i, mib[2];
        size_t size;

        mib[0] = CTL_VM;
        mib[1] = VM_LOADAVG;
        size = sizeof(loadinfo);
        if (sysctl(mib, 2, &loadinfo, &size, 0, 0) < 0)
                return (-1);

        kvm_getprocs(kd, KERN_PROC_PROC, 0, &nproc);
        kvm_getprocs(kd, KERN_PROC_ALL,  0, &nthrd);

        for (i = 0; i < 3; i++)
                loadavg[i] = (double) loadinfo.ldavg[i] / loadinfo.fscale;
        return(0);
}

double getloadavg1() {
    return avenrun[0];
}

double getloadavg5() {
    return avenrun[1];
}

double getloadavg15() {
    return avenrun[2];
}

int getnproc() {
    return nproc;
}

int getnthrd() {
    return nthrd;
}

void loadavginit() {
    kd = kvm_open(NULL, _PATH_DEVNULL, NULL, O_RDONLY, "kvm_open");
}

void loadavgfetch() {
    mygetloadavg(avenrun);
}

*/
import "C"

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"
	"github.com/blabber/go-freebsd-sysctl/sysctl"
)

const (
	// VENDOR namespace part
	pluginVendor = "swills"
	// FS namespace part
	fs = "fbsd"
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
	C.loadavginit()
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

	// read metrics from provided loadavg loacation
	if err := getStats(&stats, lp.cpus); err != nil {
		lp.logger.Errorf("Could not read metrics", err)
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
	var ncpu int64
	var err error
	ncpu, err = sysctl.GetInt64("hw.ncpu")
	if err != nil {
		return -1, err
	}
	return int(ncpu), nil
}

func getStats(stats *LoadMetrics, cpus int) error {
	C.loadavgfetch()
	stats.Min1 = float64(C.getloadavg1())
	stats.Min1Rel = stats.Min1 / float64(cpus)
	stats.Min5 = float64(C.getloadavg5())
	stats.Min5Rel = stats.Min5 / float64(cpus)
	stats.Min15 = float64(C.getloadavg15())
	stats.Min15Rel = stats.Min15 / float64(cpus)
	stats.RunSched = int(C.getnproc())
	stats.ExistingSched = int(C.getnthrd())
	return nil
}

func getInfoFields(metric string) infoFields {
	info, ok := pluginInfoFields[metric]
	if !ok {
		info = infoFields{description: "", unit: ""}
	}
	return info
}
