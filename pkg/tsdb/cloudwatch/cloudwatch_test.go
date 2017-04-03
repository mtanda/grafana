package cloudwatch

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCloudWatch(t *testing.T) {
	Convey("CloudWatch", t, func() {

		Convey("can parse cloudwatch json model", func() {
			json := `
        {
              "region": "us-east-1",
              "namespace": "AWS/ApplicationELB",
              "metricName": "TargetResponseTime",
              "dimensions": {
                "LoadBalancer": "lb",
                "TargetGroup": "tg"
              },
              "statistics": [
                "Average",
                "Maximum",
                "p50.00",
                "p90.00"
              ],
              "period": "60"
      }
      `
			modelJson, err := simplejson.NewJson([]byte(json))
			So(err, ShouldBeNil)

			res, err := parseQuery(modelJson)
			So(err, ShouldBeNil)
			So(res.Region, ShouldEqual, "us-east-1")
			So(res.Namespace, ShouldEqual, "AWS/ApplicationELB")
			So(res.MetricName, ShouldEqual, "TargetResponseTime")
			So(len(res.Dimensions), ShouldEqual, 2)
			So(*res.Dimensions[0].Name, ShouldEqual, "LoadBalancer")
			So(*res.Dimensions[0].Value, ShouldEqual, "lb")
			So(*res.Dimensions[1].Name, ShouldEqual, "TargetGroup")
			So(*res.Dimensions[1].Value, ShouldEqual, "tg")
			So(len(res.Statistics), ShouldEqual, 2)
			So(*res.Statistics[0], ShouldEqual, "Average")
			So(*res.Statistics[1], ShouldEqual, "Maximum")
			So(len(res.ExtendedStatistics), ShouldEqual, 2)
			So(*res.ExtendedStatistics[0], ShouldEqual, "p50.00")
			So(*res.ExtendedStatistics[1], ShouldEqual, "p90.00")
			So(res.Period, ShouldEqual, 60)
		})
	})
}
