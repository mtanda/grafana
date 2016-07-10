define([
  'angular'
],
function (angular) {
  'use strict';

  var module = angular.module('grafana.services');

  module.service('sharedSrv', function($q) {

    this.getSharedObject = function() {
      return $q.when(
        {
          foo: {
            bar: 'baz'
          }
        }
      );
    };
  });
});
