define([
  'lodash',
  'moment',
],
function (_, moment) {
  'use strict';

  function PrometheusMetricFindQuery(datasource, query) {
    this.datasource = datasource;
    this.query = query;
  }

  PrometheusMetricFindQuery.prototype.process = function() {
    var label_values_regex = /^label_values\(([^,]+)(?:,\s*(.+))?\)$/;
    var metric_names_regex = /^metrics\((.+)\)$/;
    var query_result_regex = /^query_result\((.+)\)$/;

    var url;
    var label_values_query = this.query.match(label_values_regex);
    if (label_values_query) {
      if (!label_values_query[2]) {
        // return label values globally
        url = '/api/v1/label/' + label_values_query[1] + '/values';

        return this.datasource._request('GET', url).then(function(result) {
          return _.map(result.data.data, function(value) {
            return {text: value};
          });
        });
      } else {
        url = '/api/v1/series?match[]=' + encodeURIComponent(label_values_query[1]);

        return this.datasource._request('GET', url)
        .then(function(result) {
          return _.map(result.data.data, function(metric) {
            return {
              text: metric[label_values_query[2]],
              expandable: true
            };
          });
        });
      }
    }

    var metric_names_query = this.query.match(metric_names_regex);
    if (metric_names_query) {
      url = '/api/v1/label/__name__/values';

      return this.datasource._request('GET', url)
      .then(function(result) {
        return _.chain(result.data.data)
        .filter(function(metricName) {
          var r = new RegExp(metric_names_query[1]);
          return r.test(metricName);
        })
        .map(function(matchedMetricName) {
          return {
            text: matchedMetricName,
            expandable: true
          };
        })
        .value();
      });
    }

    var query_result_query = this.query.match(query_result_regex);
    if (query_result_query) {
      url = '/api/v1/query?query=' + encodeURIComponent(query_result_query[1]) + '&time=' + (moment().valueOf() / 1000);

      return this.datasource._request('GET', url)
      .then(function(result) {
        return _.map(result.data.data.result, function(metricData) {
          var text = metricData.metric.__name__ || '';
          delete metricData.metric.__name__;
          text += '{' +
                  _.map(metricData.metric, function(v, k) { return k + '="' + v + '"'; }).join(',') +
                  '}';
          text += ' ' + metricData.value[1] + ' ' + metricData.value[0] * 1000;

          return {
            text: text,
            expandable: true
          };
        });
      });
    }

    // if query contains full metric name, return metric name and label list
    url = '/api/v1/series?match[]=' + encodeURIComponent(this.query);

    return this.datasource._request('GET', url)
    .then(function(result) {
      return _.map(result.data.data, function(metric) {
        return {
          text: getOriginalMetricName(metric),
          expandable: true
        };
      });
    });
  };

  function getOriginalMetricName(labelData) {
    var metricName = labelData.__name__ || '';
    delete labelData.__name__;
    var labelPart = _.map(_.pairs(labelData), function(label) {
      return label[0] + '="' + label[1] + '"';
    }).join(',');
    return metricName + '{' + labelPart + '}';
  }

  return PrometheusMetricFindQuery;
});
