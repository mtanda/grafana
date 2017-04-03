package cloudwatch

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type CloudWatchQuery struct {
	Region             string
	Namespace          string
	MetricName         string
	Dimensions         []*cloudwatch.Dimension
	Statistics         []*string
	ExtendedStatistics []*string
	Period             int
	StartTime          time.Time
	EndTime            time.Time
}
