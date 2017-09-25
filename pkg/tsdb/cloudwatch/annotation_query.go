package cloudwatch

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/tsdb"
)

func (e *CloudWatchExecutor) executeAnnotationQuery(ctx context.Context, queryContext *tsdb.TsdbQuery) (*tsdb.Response, error) {
	result := &tsdb.Response{
		Results: make(map[string]*tsdb.QueryResult),
	}
	firstQuery := queryContext.Queries[0]
	queryResult := &tsdb.QueryResult{Meta: simplejson.New(), RefId: firstQuery.RefId}

	parameters := firstQuery.Model
	usePrefixMatch := parameters.Get("prefixMatching").MustBool()
	region := parameters.Get("region").MustString("")
	namespace := parameters.Get("namespace").MustString("")
	metricName := parameters.Get("metricName").MustString("")
	dimensions := parameters.Get("dimensions").MustMap()
	statistics := parameters.Get("statistics").MustArray()
	extendedStatistics := parameters.Get("extendedStatistics").MustArray()
	period := 300
	if usePrefixMatch {
		period := parameters.Get("period").MustInt()
	}
	actionPrefix := parameters.Get("actionPrefix").MustString("")
	alarmNamePrefix := parameters.Get("alarmNamePrefix").MustString("")

	dsInfo := e.getDsInfo(region)
	cfg, err := getAwsConfig(dsInfo)
	if err != nil {
		return nil, errors.New("Failed to call cloudwatch:ListMetrics")
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, errors.New("Failed to call cloudwatch:ListMetrics")
	}
	svc := cloudwatch.New(sess, cfg)

	//alarms var
	if usePrefixMatch {
		params := &cloudwatch.DescribeAlarmsInput{
			MaxRecords:      aws.Int64(100),
			ActionPrefix:    aws.String(actionPrefix),
			AlarmNamePrefix: aws.String(alarmNamePrefix),
		}
		resp, err := svc.DescribeAlarms(params)
	} else {
		if region == "" || namespace == "" || metricName == "" || len(statistics) == 0 {
			return result, nil
		}

		params := &cloudwatch.DescribeAlarmsForMetricInput{
			Namespace:         aws.String(namespace),
			MetricName:        aws.String(metricName),
			Period:            aws.Int64(int64(period)),
			Dimensions:        dimensions,
			Statistic:         statistic,
			ExtendedStatistic: extendedStatistic,
		}
		resp, err := svc.DescribeAlarmsForMetric(params)
	}

	transformToTable(data, queryResult)
	result.Results[firstQuery.RefId] = queryResult
	return result, err
}

/*

    var allQueryPromise;
    if (usePrefixMatch) {
      allQueryPromise = [
        this.datasource.performDescribeAlarms(region, actionPrefix, alarmNamePrefix, [], '').then(function(alarms) {
          alarms.MetricAlarms = self.filterAlarms(alarms, namespace, metricName, dimensions, statistics, period);
          return alarms;
        })
      ];
    } else {
      if (!region || !namespace || !metricName || _.isEmpty(statistics)) { return this.$q.when([]); }

      allQueryPromise = _.map(statistics, function(statistic) {
        return self.datasource.performDescribeAlarmsForMetric(region, namespace, metricName, dimensions, statistic, period);
      });
    }
    this.$q.all(allQueryPromise).then(function(alarms) {
      var eventList = [];

      var start = self.datasource.convertToCloudWatchTime(from, false);
      var end = self.datasource.convertToCloudWatchTime(to, true);
      _.chain(alarms)
      .map('MetricAlarms')
      .flatten()
      .each(function(alarm) {
        if (!alarm) {
          d.resolve(eventList);
          return;
        }

        self.datasource.performDescribeAlarmHistory(region, alarm.AlarmName, start, end).then(function(history) {
          _.each(history.AlarmHistoryItems, function(h) {
            var event = {
              annotation: self.annotation,
              time: Date.parse(h.Timestamp),
              title: h.AlarmName,
              tags: [h.HistoryItemType],
              text: h.HistorySummary
            };

            eventList.push(event);
          });

          d.resolve(eventList);
        });
      })
      .value();
    });

    return d.promise;
  };

  CloudWatchAnnotationQuery.prototype.filterAlarms = function(alarms, namespace, metricName, dimensions, statistics, period) {
    return _.filter(alarms.MetricAlarms, function(alarm) {
      if (!_.isEmpty(namespace) && alarm.Namespace !== namespace) {
        return false;
      }
      if (!_.isEmpty(metricName) && alarm.MetricName !== metricName) {
        return false;
      }
      var sd = function(d) {
        return d.Name;
      };
      var isSameDimensions = JSON.stringify(_.sortBy(alarm.Dimensions, sd)) === JSON.stringify(_.sortBy(dimensions, sd));
      if (!_.isEmpty(dimensions) && !isSameDimensions) {
        return false;
      }
      if (!_.isEmpty(statistics) && !_.includes(statistics, alarm.Statistic)) {
        return false;
      }
      if (!_.isNaN(period) && alarm.Period !== period) {
        return false;
      }
      return true;
    });
  };

  return CloudWatchAnnotationQuery;
});

*/
