package cloudwatch

import (
	"testing"

	p "github.com/cloudwatch/common/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCloudWatch(t *testing.T) {
	Convey("CloudWatch", t, func() {

		Convey("converting metric name", func() {
			metric := map[p.LabelName]p.LabelValue{
				p.LabelName("app"):    p.LabelValue("backend"),
				p.LabelName("device"): p.LabelValue("mobile"),
			}

			query := &CloudWatchQuery{
				LegendFormat: "legend {{app}} {{ device }} {{broken}}",
			}

			So(formatLegend(metric, query), ShouldEqual, "legend backend mobile {{broken}}")
		})

		Convey("build full serie name", func() {
			metric := map[p.LabelName]p.LabelValue{
				p.LabelName(p.MetricNameLabel): p.LabelValue("http_request_total"),
				p.LabelName("app"):             p.LabelValue("backend"),
				p.LabelName("device"):          p.LabelValue("mobile"),
			}

			query := &CloudWatchQuery{
				LegendFormat: "",
			}

			So(formatLegend(metric, query), ShouldEqual, `http_request_total{app="backend", device="mobile"}`)
		})
	})
}
