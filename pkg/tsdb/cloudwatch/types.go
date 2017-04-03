package cloudwatch

import "time"

type CloudWatchQuery struct {
	Expr         string
	Step         time.Duration
	LegendFormat string
	Start        time.Time
	End          time.Time
}
