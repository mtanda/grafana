package cloudwatch

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/grafana/grafana/pkg/middleware"
)

type actionHandler func(*cwRequest, *middleware.Context)

var actionHandlers map[string]actionHandler

type cwRequest struct {
	Region string `json:"region"`
	Action string `json:"action"`
	Body   []byte `json:"-"`
}

func init() {
	actionHandlers = map[string]actionHandler{
		"GetMetricStatistics":     handleGetMetricStatistics,
		"ListMetrics":             handleListMetrics,
		"DescribeAlarmsForMetric": handleDescribeAlarmsForMetric,
		"DescribeAlarmHistory":    handleDescribeAlarmHistory,
		"DescribeInstances":       handleDescribeInstances,
		"__GetRegions":            handleGetRegions,
		"__GetNamespaces":         handleGetNamespaces,
		"__GetMetrics":            handleGetMetrics,
		"__GetDimensions":         handleGetDimensions,
	}
	mockMode := true
	if mockMode {
		actionHandlers["GetMetricStatistics"] = handleMockGetMetricStatistics
		actionHandlers["ListMetrics"] = handleMockListMetrics
	}
}

func handleGetMetricStatistics(req *cwRequest, c *middleware.Context) {
	svc := cloudwatch.New(&aws.Config{Region: aws.String(req.Region)})

	reqParam := &struct {
		Parameters struct {
			Namespace  string                  `json:"namespace"`
			MetricName string                  `json:"metricName"`
			Dimensions []*cloudwatch.Dimension `json:"dimensions"`
			Statistics []*string               `json:"statistics"`
			StartTime  int64                   `json:"startTime"`
			EndTime    int64                   `json:"endTime"`
			Period     int64                   `json:"period"`
		} `json:"parameters"`
	}{}
	json.Unmarshal(req.Body, reqParam)

	params := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(reqParam.Parameters.Namespace),
		MetricName: aws.String(reqParam.Parameters.MetricName),
		Dimensions: reqParam.Parameters.Dimensions,
		Statistics: reqParam.Parameters.Statistics,
		StartTime:  aws.Time(time.Unix(reqParam.Parameters.StartTime, 0)),
		EndTime:    aws.Time(time.Unix(reqParam.Parameters.EndTime, 0)),
		Period:     aws.Int64(reqParam.Parameters.Period),
	}

	resp, err := svc.GetMetricStatistics(params)
	if err != nil {
		c.JsonApiErr(500, "Unable to call AWS API", err)
		return
	}

	c.JSON(200, resp)
}

func handleMockGetMetricStatistics(req *cwRequest, c *middleware.Context) {
	reqParam := &struct {
		Parameters struct {
			Namespace  string                  `json:"namespace"`
			MetricName string                  `json:"metricName"`
			Dimensions []*cloudwatch.Dimension `json:"dimensions"`
			Statistics []*string               `json:"statistics"`
			StartTime  int64                   `json:"startTime"`
			EndTime    int64                   `json:"endTime"`
			Period     int64                   `json:"period"`
		} `json:"parameters"`
	}{}
	json.Unmarshal(req.Body, reqParam)

	label := reqParam.Parameters.MetricName

	n := (reqParam.Parameters.EndTime - reqParam.Parameters.StartTime) / reqParam.Parameters.Period
	dps := make([]cloudwatch.Datapoint, n)
	for i, dp := range dps {
		for _, s := range reqParam.Parameters.Statistics {
			switch
			dps[i].s = 1
		}
	}

	//resp =
	//  Datapoints: [
	//    {
	//      Average: 1,
	//      Timestamp: 'Wed Dec 31 1969 16:00:00 GMT-0800 (PST)'
	//    }
	//  ],
	//  Label: 'CPUUtilization'
}

func handleListMetrics(req *cwRequest, c *middleware.Context) {
	svc := cloudwatch.New(&aws.Config{Region: aws.String(req.Region)})
	reqParam := &struct {
		Parameters struct {
			Namespace  string                        `json:"namespace"`
			MetricName string                        `json:"metricName"`
			Dimensions []*cloudwatch.DimensionFilter `json:"dimensions"`
		} `json:"parameters"`
	}{}

	json.Unmarshal(req.Body, reqParam)

	params := &cloudwatch.ListMetricsInput{
		Namespace:  aws.String(reqParam.Parameters.Namespace),
		MetricName: aws.String(reqParam.Parameters.MetricName),
		Dimensions: reqParam.Parameters.Dimensions,
	}

	resp, err := svc.ListMetrics(params)
	if err != nil {
		c.JsonApiErr(500, "Unable to call AWS API", err)
		return
	}

	c.JSON(200, resp)
}

func handleDescribeAlarmsForMetric(req *cwRequest, c *middleware.Context) {
	svc := cloudwatch.New(&aws.Config{Region: aws.String(req.Region)})
	reqParam := &struct {
		Parameters struct {
			Namespace  string                  `json:"namespace"`
			MetricName string                  `json:"metricName"`
			Dimensions []*cloudwatch.Dimension `json:"dimensions"`
			Statistic  string                  `json:"statistic"`
			Period     int64                   `json:"period"`
		} `json:"parameters"`
	}{}
	json.Unmarshal(req.Body, reqParam)

	params := &cloudwatch.DescribeAlarmsForMetricInput{
		Namespace:  aws.String(reqParam.Parameters.Namespace),
		MetricName: aws.String(reqParam.Parameters.MetricName),
		Period:     aws.Int64(reqParam.Parameters.Period),
	}
	if len(reqParam.Parameters.Dimensions) != 0 {
		params.Dimensions = reqParam.Parameters.Dimensions
	}
	if reqParam.Parameters.Statistic != "" {
		params.Statistic = aws.String(reqParam.Parameters.Statistic)
	}

	resp, err := svc.DescribeAlarmsForMetric(params)
	if err != nil {
		c.JsonApiErr(500, "Unable to call AWS API", err)
		return
	}

	c.JSON(200, resp)
}

func handleDescribeAlarmHistory(req *cwRequest, c *middleware.Context) {
	svc := cloudwatch.New(&aws.Config{Region: aws.String(req.Region)})
	reqParam := &struct {
		Parameters struct {
			AlarmName       string `json:"alarmName"`
			HistoryItemType string `json:"historyItemType"`
			StartDate       int64  `json:"startDate"`
			EndDate         int64  `json:"endDate"`
		} `json:"parameters"`
	}{}
	json.Unmarshal(req.Body, reqParam)

	params := &cloudwatch.DescribeAlarmHistoryInput{
		AlarmName: aws.String(reqParam.Parameters.AlarmName),
		StartDate: aws.Time(time.Unix(reqParam.Parameters.StartDate, 0)),
		EndDate:   aws.Time(time.Unix(reqParam.Parameters.EndDate, 0)),
	}
	if reqParam.Parameters.HistoryItemType != "" {
		params.HistoryItemType = aws.String(reqParam.Parameters.HistoryItemType)
	}

	resp, err := svc.DescribeAlarmHistory(params)
	if err != nil {
		c.JsonApiErr(500, "Unable to call AWS API", err)
		return
	}

	c.JSON(200, resp)
}

func handleDescribeInstances(req *cwRequest, c *middleware.Context) {
	svc := ec2.New(&aws.Config{Region: aws.String(req.Region)})

	reqParam := &struct {
		Parameters struct {
			Filters     []*ec2.Filter `json:"filters"`
			InstanceIds []*string     `json:"instanceIds"`
		} `json:"parameters"`
	}{}
	json.Unmarshal(req.Body, reqParam)

	params := &ec2.DescribeInstancesInput{}
	if len(reqParam.Parameters.Filters) > 0 {
		params.Filters = reqParam.Parameters.Filters
	}
	if len(reqParam.Parameters.InstanceIds) > 0 {
		params.InstanceIds = reqParam.Parameters.InstanceIds
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		c.JsonApiErr(500, "Unable to call AWS API", err)
		return
	}

	c.JSON(200, resp)
}

func HandleRequest(c *middleware.Context) {
	var req cwRequest
	req.Body, _ = ioutil.ReadAll(c.Req.Request.Body)
	json.Unmarshal(req.Body, &req)

	if handler, found := actionHandlers[req.Action]; !found {
		c.JsonApiErr(500, "Unexpected AWS Action", errors.New(req.Action))
		return
	} else {
		handler(&req, c)
	}
}
