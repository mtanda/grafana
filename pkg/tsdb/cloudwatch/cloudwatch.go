package cloudwatch

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/tsdb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	cwapi "github.com/grafana/grafana/pkg/api/cloudwatch"
	"github.com/grafana/grafana/pkg/components/null"
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

	query, err := parseQuery(queries[0].Model)
	if err != nil {
		return result.WithError(err)
	}

	client, err := e.getClient(query.Region)
	if err != nil {
		return result.WithError(err)
	}

	startTime, err := queryContext.TimeRange.ParseFrom()
	if err != nil {
		return result.WithError(err)
	}

	endTime, err := queryContext.TimeRange.ParseTo()
	if err != nil {
		return result.WithError(err)
	}

	params := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(query.Namespace),
		MetricName: aws.String(query.MetricName),
		Dimensions: query.Dimensions,
		Period:     aws.Int64(int64(query.Period)),
		StartTime:  aws.Time(startTime.Add(-time.Minute * 10)),
		EndTime:    aws.Time(endTime.Add(-time.Minute * 10)),
	}
	if len(query.Statistics) > 0 {
		params.Statistics = query.Statistics
	}
	if len(query.ExtendedStatistics) > 0 {
		params.ExtendedStatistics = query.ExtendedStatistics
	}

	resp, err := client.GetMetricStatistics(params)
	if err != nil {
		return result.WithError(err)
	}

	queryResult, err := parseResponse(resp, query)
	if err != nil {
		return result.WithError(err)
	}

	result.QueryResults = queryResult
	return result
}

func (e *CloudWatchExecutor) getClient(region string) (*cloudwatch.CloudWatch, error) {
	assumeRoleArn := e.DataSource.JsonData.Get("assumeRoleArn").MustString()

	accessKey := ""
	secretKey := ""
	for key, value := range e.DataSource.SecureJsonData.Decrypt() {
		if key == "accessKey" {
			accessKey = value
		}
		if key == "secretKey" {
			secretKey = value
		}
	}

	datasourceInfo := &cwapi.DatasourceInfo{
		Region:        region,
		Profile:       e.DataSource.Database,
		AssumeRoleArn: assumeRoleArn,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
	}

	credentials, err := cwapi.GetCredentials(datasourceInfo)
	if err != nil {
		return nil, err
	}

	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials,
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	client := cloudwatch.New(sess, cfg)
	return client, nil
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
	if p == "" {
		if namespace == "AWS/EC2" {
			p = "300"
		} else {
			p = "60"
		}
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

func parseResponse(resp *cloudwatch.GetMetricStatisticsOutput, query *CloudWatchQuery) (map[string]*tsdb.QueryResult, error) {
	queryResults := make(map[string]*tsdb.QueryResult)
	queryRes := tsdb.NewQueryResult()

	var value float64
	for _, s := range append(query.Statistics, query.ExtendedStatistics...) {
		series := tsdb.TimeSeries{
			Name: *resp.Label,
			Tags: map[string]string{},
		}

		for _, d := range query.Dimensions {
			series.Tags[*d.Name] = *d.Value
		}

		for _, v := range resp.Datapoints {
			switch *s {
			case "Average":
				value = *v.Average
			case "Maximum":
				value = *v.Maximum
			case "Minimum":
				value = *v.Minimum
			case "Sum":
				value = *v.Sum
			case "SampleCount":
				value = *v.SampleCount
			default:
				if strings.Index(*s, "p") == 0 && v.ExtendedStatistics[*s] != nil {
					value = *v.ExtendedStatistics[*s]
				}
			}
			series.Points = append(series.Points, tsdb.NewTimePoint(null.FloatFrom(value), float64(v.Timestamp.Unix()*1000)))
		}

		queryRes.Series = append(queryRes.Series, &series)
	}

	queryResults["A"] = queryRes
	return queryResults, nil
}
