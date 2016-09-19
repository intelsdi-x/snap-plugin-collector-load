// +build small

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
	"os"
	"testing"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LoadInfoSuite struct {
	suite.Suite
	MockLoadInfo      string
	min1, min5, min15 float64
	runsch, existsch  int
	sch               string
}

func (lis *LoadInfoSuite) SetupSuite() {
	lis.min1 = 0.40
	lis.min5 = 0.08
	lis.min15 = 0.16
	lis.runsch = 1
	lis.existsch = 100
	lis.MockLoadInfo = "."
	createMockLoadInfo("loadavg", lis.min1, lis.min5, lis.min15, lis.runsch, lis.existsch, 1111)
}

func (lis *LoadInfoSuite) TearDownSuite() {
	removeMockLoadInfo("loadavg")
}

func (lis *LoadInfoSuite) TestGetStats() {
	Convey("Given load info map", lis.T(), func() {
		stats := LoadMetrics{}

		Convey("and mock memory info file created", func() {
			assert.Equal(lis.T(), ".", lis.MockLoadInfo)
		})

		Convey("When reading load statistics from file", func() {
			err := getStats(lis.MockLoadInfo, &stats, 2)

			Convey("No error should be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Proper statistics values are returned", func() {
				So(stats.Min1, ShouldEqual, lis.min1)
				So(stats.Min1Rel, ShouldEqual, lis.min1/float64(2))
				So(stats.Min5, ShouldEqual, lis.min5)
				So(stats.Min5Rel, ShouldEqual, lis.min5/float64(2))
				So(stats.Min15, ShouldEqual, lis.min15)
				So(stats.Min15Rel, ShouldEqual, lis.min15/float64(2))
				So(stats.RunSched, ShouldEqual, lis.runsch)
				So(stats.ExistingSched, ShouldEqual, lis.existsch)
			})

		})
	})
}

func (lis *LoadInfoSuite) TestGetMetricTypes() {
	_ = plugin.ConfigType{}
	Convey("Given load info plugin initialized", lis.T(), func() {
		loadPlg := New()

		Convey("When one wants to get list of available meterics", func() {
			mts, err := loadPlg.GetMetricTypes(plugin.ConfigType{})

			Convey("Then error should not be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then list of metrics is returned", func() {
				So(len(mts), ShouldEqual, 8)

				namespaces := []string{}
				for _, m := range mts {
					namespaces = append(namespaces, m.Namespace().String())
				}

				So(namespaces, ShouldContain, "/intel/procfs/load/min1")
				So(namespaces, ShouldContain, "/intel/procfs/load/min1_rel")
				So(namespaces, ShouldContain, "/intel/procfs/load/min5")
				So(namespaces, ShouldContain, "/intel/procfs/load/min5_rel")
				So(namespaces, ShouldContain, "/intel/procfs/load/min15")
				So(namespaces, ShouldContain, "/intel/procfs/load/min15_rel")
				So(namespaces, ShouldContain, "/intel/procfs/load/runnable_scheduling")
				So(namespaces, ShouldContain, "/intel/procfs/load/existing_scheduling")
			})
		})
	})
}

func (lis *LoadInfoSuite) TestCollectMetrics() {
	Convey("Given memInfo plugin initlialized", lis.T(), func() {
		loadPlg := New()

		Convey("When one wants to get values for given metric types", func() {
			cfg := plugin.NewPluginConfigType()
			cfg.AddItem("proc_path", ctypes.ConfigValueStr{lis.MockLoadInfo})
			mTypes := []plugin.MetricType{
				plugin.MetricType{Namespace_: core.NewNamespace("intel", "procfs", "load", "min1"), Config_: cfg.ConfigDataNode},
				plugin.MetricType{Namespace_: core.NewNamespace("intel", "procfs", "load", "runnable_scheduling"), Config_: cfg.ConfigDataNode},
				plugin.MetricType{Namespace_: core.NewNamespace("intel", "procfs", "load", "min15"), Config_: cfg.ConfigDataNode},
				plugin.MetricType{Namespace_: core.NewNamespace("intel", "procfs", "load", "min15_rel"), Config_: cfg.ConfigDataNode},
			}

			metrics, err := loadPlg.CollectMetrics(mTypes)

			Convey("Then no erros should be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then proper metrics values are returned", func() {
				So(len(metrics), ShouldEqual, 4)

				stats := map[string]interface{}{}
				for _, m := range metrics {
					n := m.Namespace().String()
					stats[n] = m.Data()
				}

				So(len(metrics), ShouldEqual, len(stats))

				So(stats["/intel/procfs/load/min1"], ShouldNotBeNil)
				So(stats["/intel/procfs/load/runnable_scheduling"], ShouldNotBeNil)
				So(stats["/intel/procfs/load/min15"], ShouldNotBeNil)
				So(stats["/intel/procfs/load/min15_rel"], ShouldNotBeNil)
			})
		})
	})
}

func TestGetStatsSuite(t *testing.T) {
	suite.Run(t, &LoadInfoSuite{MockLoadInfo: "mockLoadInfo"})
}

func createMockLoadInfo(loadInfo string, min1 float64, min5 float64, min15 float64, runsch int, existsch int, pid uint) {
	content := fmt.Sprintf(
		"%f %f %f %d/%d %d",
		min1, min5, min15, runsch, existsch, pid)
	loadInfoContent := []byte(content)
	f, err := os.Create(loadInfo)
	if err != nil {
		panic(err)
	}
	f.Write(loadInfoContent)
}

func removeMockLoadInfo(loadInfo string) {
	os.Remove(loadInfo)
}
