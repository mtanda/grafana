define([
  'lodash',
  'moment',
],
function (_, moment) {
  'use strict';

  function CloudWatchAnnotationQuery(datasource, query) {
    this.datasource = datasource;
    this.query = query;
  }

  CloudWatchAnnotationQuery.prototype.process = function() {
    var usePrefixMatch = annotation.prefixMatching;
    var region = templateSrv.replace(annotation.region);
    var namespace = templateSrv.replace(annotation.namespace);
    var metricName = templateSrv.replace(annotation.metricName);
    var dimensions = convertDimensionFormat(annotation.dimensions);
    var statistics = _.map(annotation.statistics, function(s) { return templateSrv.replace(s); });
    var defaultPeriod = usePrefixMatch ? '' : '300';
    var period = annotation.period || defaultPeriod;
    period = parseInt(period, 10);
    var actionPrefix = annotation.actionPrefix || '';
    var alarmNamePrefix = annotation.alarmNamePrefix || '';

    var d = $q.defer();
    var self = this;
    var allQueryPromise;
    if (usePrefixMatch) {
      allQueryPromise = [
        this.performDescribeAlarms(region, actionPrefix, alarmNamePrefix, [], '').then(function(alarms) {
          alarms.MetricAlarms = _.filter(alarms.MetricAlarms, function(alarm) {
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
            if (!_.isEmpty(statistics) && !_.contains(statistics, alarm.Statistic)) {
              return false;
            }
            if (!_.isNaN(period) && alarm.Period !== period) {
              return false;
            }
            return true;
          });

          return alarms;
        })
      ];
    } else {
      if (!region || !namespace || !metricName || _.isEmpty(statistics)) { return $q.when([]); }

      allQueryPromise = _.map(statistics, function(statistic) {
        return self.performDescribeAlarmsForMetric(region, namespace, metricName, dimensions, statistic, period);
      });
    }
    $q.all(allQueryPromise).then(function(alarms) {
      var eventList = [];

      var start = convertToCloudWatchTime(options.range.from, false);
      var end = convertToCloudWatchTime(options.range.to, true);
      _.chain(alarms)
      .pluck('MetricAlarms')
      .flatten()
      .each(function(alarm) {
        if (!alarm) {
          d.resolve(eventList);
          return;
        }

        self.performDescribeAlarmHistory(region, alarm.AlarmName, start, end).then(function(history) {
          _.each(history.AlarmHistoryItems, function(h) {
            var event = {
              annotation: annotation,
              time: Date.parse(h.Timestamp),
              title: h.AlarmName,
              tags: [h.HistoryItemType],
              text: h.HistorySummary
            };

            eventList.push(event);
          });

          d.resolve(eventList);
        });
      });
    });

    return d.promise;
  };

  return CloudWatchAnnotationQuery;
});
