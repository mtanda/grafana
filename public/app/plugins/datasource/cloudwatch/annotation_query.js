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
  };

  return CloudWatchAnnotationQuery;
});
