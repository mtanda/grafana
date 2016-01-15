define([
  'lodash',
  'moment',
],
function (_, moment) {
  'use strict';

  function CloudWatchAnnotationQuery(datasource, annotation, $q, templateSrv) {
    this.datasource = datasource;
    this.annotation = annotation;
    this.$q = $q;
    this.templateSrv = templateSrv;
  }

  CloudWatchAnnotationQuery.prototype.process = function(from, to) {
    var usePrefixMatch = this.annotation.prefixMatching;
    var region = this.templateSrv.replace(this.annotation.region);
    var namespace = this.templateSrv.replace(this.annotation.namespace);
    var metricName = this.templateSrv.replace(this.annotation.metricName);
    var dimensions = this.datasource.convertDimensionFormat(this.annotation.dimensions);
    var statistics = _.map(this.annotation.statistics, function(s) { return this.templateSrv.replace(s); });
    var defaultPeriod = usePrefixMatch ? '' : '300';
    var period = this.annotation.period || defaultPeriod;
    period = parseInt(period, 10);
    var actionPrefix = this.annotation.actionPrefix || '';
    var alarmNamePrefix = this.annotation.alarmNamePrefix || '';

    var d = this.$q.defer();
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
      if (!region || !namespace || !metricName || _.isEmpty(statistics)) { return this.$q.when([]); }

      allQueryPromise = _.map(statistics, function(statistic) {
        return self.datasource.performDescribeAlarmsForMetric(region, namespace, metricName, dimensions, statistic, period);
      });
    }
    this.$q.all(allQueryPromise).then(function(alarms) {
      var eventList = [];

      var start = this.datasource.convertToCloudWatchTime(from, false);
      var end = this.datasoure.convertToCloudWatchTime(to, true);
      _.chain(alarms)
      .pluck('MetricAlarms')
      .flatten()
      .each(function(alarm) {
        if (!alarm) {
          d.resolve(eventList);
          return;
        }

        self.datasource.performDescribeAlarmHistory(region, alarm.AlarmName, start, end).then(function(history) {
          _.each(history.AlarmHistoryItems, function(h) {
            var event = {
              annotation: this.annotation,
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


  CloudWatchAnnotationQuery.prototype.filterAlarms = function(alarms) {
  };

  return CloudWatchAnnotationQuery;
});
