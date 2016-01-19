// +build unit

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
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/intelsdi-x/snap/control/plugin"
)

type LoadInfoSuite struct {
	suite.Suite
	MockLoadInfo      string
	min1, min5, min15 float64
	sch               string
}

func (lis *LoadInfoSuite) SetupSuite() {
	loadInfo = lis.MockLoadInfo
	lis.min1 = 0.40
	lis.min5 = 0.08
	lis.min15 = 0.16
	lis.sch = "1/100"
	createMockLoadInfo(lis.min1, lis.min5, lis.min15, lis.sch, 1111)
}

func (lis *LoadInfoSuite) TearDownSuite() {
	removeMockLoadInfo()
}

func (lis *LoadInfoSuite) TestGetStats() {
	Convey("Given load info map", lis.T(), func() {
		stats := map[string]interface{}{}

		Convey("and mock memory info file created", func() {
			assert.Equal(lis.T(), "mockLoadInfo", loadInfo)
		})

		Convey("When reading load statistics from file", func() {
			err := getStats(stats, 2)

			Convey("No error should be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Proper statistics values are returned", func() {
				val, ok := stats["min1"].(float64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, lis.min1)

				val, ok = stats["min1_rel"].(float64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, lis.min1/float64(2))

				val, ok = stats["min5"].(float64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, lis.min5)

				val, ok = stats["min5_rel"].(float64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, lis.min5/float64(2))

				val, ok = stats["min15"].(float64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, lis.min15)

				val, ok = stats["min15_rel"].(float64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, lis.min15/float64(2))

				sch, ok := stats["scheduling"].(string)
				So(ok, ShouldBeTrue)
				So(sch, ShouldEqual, lis.sch)
			})

		})
	})
}

func (lis *LoadInfoSuite) TestGetMetricTypes() {
	_ = plugin.PluginConfigType{}
	Convey("Given load info plugin initialized", lis.T(), func() {
		loadPlg := New()

		Convey("When one wants to get list of available meterics", func() {
			mts, err := loadPlg.GetMetricTypes(plugin.PluginConfigType{})

			Convey("Then error should not be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then list of metrics is returned", func() {
				So(len(mts), ShouldEqual, 7)

				namespaces := []string{}
				for _, m := range mts {
					namespaces = append(namespaces, strings.Join(m.Namespace(), "/"))
				}

				So(namespaces, ShouldContain, "intel/procfs/load/min1")
				So(namespaces, ShouldContain, "intel/procfs/load/min1_rel")
				So(namespaces, ShouldContain, "intel/procfs/load/min5")
				So(namespaces, ShouldContain, "intel/procfs/load/min5_rel")
				So(namespaces, ShouldContain, "intel/procfs/load/min15")
				So(namespaces, ShouldContain, "intel/procfs/load/min15_rel")
				So(namespaces, ShouldContain, "intel/procfs/load/scheduling")
			})
		})
	})
}

func (lis *LoadInfoSuite) TestCollectMetrics() {
	Convey("Given memInfo plugin initlialized", lis.T(), func() {
		loadPlg := New()

		Convey("When one wants to get values for given metric types", func() {
			mTypes := []plugin.PluginMetricType{
				plugin.PluginMetricType{Namespace_: []string{"intel", "procfs", "load", "min1"}},
				plugin.PluginMetricType{Namespace_: []string{"intel", "procfs", "load", "scheduling"}},
				plugin.PluginMetricType{Namespace_: []string{"intel", "procfs", "load", "min15"}},
				plugin.PluginMetricType{Namespace_: []string{"intel", "procfs", "load", "min15_rel"}},
			}

			metrics, err := loadPlg.CollectMetrics(mTypes)

			Convey("Then no erros should be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then proper metrics values are returned", func() {
				So(len(metrics), ShouldEqual, 4)

				stats := map[string]interface{}{}
				for _, m := range metrics {
					n := strings.Join(m.Namespace(), "/")
					stats[n] = m.Data()
				}

				So(len(metrics), ShouldEqual, len(stats))

				So(stats["intel/procfs/load/min1"], ShouldNotBeNil)
				So(stats["intel/procfs/load/scheduling"], ShouldNotBeNil)
				So(stats["intel/procfs/load/min15"], ShouldNotBeNil)
				So(stats["intel/procfs/load/min15_rel"], ShouldNotBeNil)
			})
		})
	})
}

func TestGetStatsSuite(t *testing.T) {
	suite.Run(t, &LoadInfoSuite{MockLoadInfo: "mockLoadInfo"})
}

func createMockLoadInfo(min1 float64, min5 float64, min15 float64, sch string, pid uint) {
	content := fmt.Sprintf(
		"%f %f %f %s %d",
		min1, min5, min15, sch, pid)
	loadInfoContent := []byte(content)
	f, err := os.Create(loadInfo)
	if err != nil {
		panic(err)
	}
	f.Write(loadInfoContent)
}

func removeMockLoadInfo() {
	os.Remove(loadInfo)
}
