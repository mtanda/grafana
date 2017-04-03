package cloudwatch

import (
	"context"
	"errors"
	"strconv"

	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/tsdb"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/grafana/grafana/pkg/components/simplejson"
)

type CloudWatchExecutor struct {
	*models.DataSource
}

func NewCloudWatchExecutor(dsInfo *models.DataSource) (tsdb.Executor, error) {
	return &CloudWatchExecutor{
		DataSource: dsInfo,
	}, nil
}

var (
	plog               log.Logger
	standardStatistics map[string]bool
)

func init() {
	plog = log.New("tsdb.cloudwatch")
	tsdb.RegisterExecutor("cloudwatch", NewCloudWatchExecutor)
	standardStatistics = map[string]bool{
		"Average":     true,
		"Maximum":     true,
		"Minimum":     true,
		"Sum":         true,
		"SampleCount": true,
	}
}

func (e *CloudWatchExecutor) Execute(ctx context.Context, queries tsdb.QuerySlice, queryContext *tsdb.QueryContext) *tsdb.BatchResult {
	result := &tsdb.BatchResult{}
	return result
}

func parseDimensions(model *simplejson.Json) ([]*cloudwatch.Dimension, error) {
	var result []*cloudwatch.Dimension

	for k, v := range model.Get("dimensions").MustMap() {
		kk := k
		if vv, ok := v.(string); ok {
			result = append(result, &cloudwatch.Dimension{
				Name:  &kk,
				Value: &vv,
			})
		} else {
			return nil, errors.New("failed to parse")
		}
	}

	return result, nil
}

func parseStatistics(model *simplejson.Json) ([]*string, []*string, error) {
	var statistics []*string
	var extendedStatistics []*string

	for _, s := range model.Get("statistics").MustArray() {
		if ss, ok := s.(string); ok {
			if _, isStandard := standardStatistics[ss]; isStandard {
				statistics = append(statistics, &ss)
			} else {
				extendedStatistics = append(extendedStatistics, &ss)
			}
		} else {
			return nil, nil, errors.New("failed to parse")
		}
	}

	return statistics, extendedStatistics, nil
}

func parseQuery(model *simplejson.Json) (*CloudWatchQuery, error) {
	region, err := model.Get("region").String()
	if err != nil {
		return nil, err
	}

	namespace, err := model.Get("namespace").String()
	if err != nil {
		return nil, err
	}

	metricName, err := model.Get("metricName").String()
	if err != nil {
		return nil, err
	}

	dimensions, err := parseDimensions(model)
	if err != nil {
		return nil, err
	}

	statistics, extendedStatistics, err := parseStatistics(model)
	if err != nil {
		return nil, err
	}

	p, err := model.Get("period").String()
	if err != nil {
		return nil, err
	}
	period, err := strconv.Atoi(p)
	if err != nil {
		return nil, err
	}

	return &CloudWatchQuery{
		Region:             region,
		Namespace:          namespace,
		MetricName:         metricName,
		Dimensions:         dimensions,
		Statistics:         statistics,
		ExtendedStatistics: extendedStatistics,
		Period:             period,
	}, nil
}
