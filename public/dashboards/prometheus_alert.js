'use strict';

// accessible variables in this scope
var ARGS, $, services;

function createGraphPanel(title, datasource, targets, span) {
  return {
    type: 'graph',
    title: title,
    datasource: datasource,
    fill: 0,
    stack: false,
    tooltip: {
      shared: true,
      value_type: 'individual'
    },
    y_formats: [
      'short',
      'short'
    ],
    targets: targets,
    span: span
  };
}

function createTextPanel(title, content, span) {
  return {
    type: 'text',
    title: title,
    mode: 'html',
    content: content,
    span: span
  };
}

function createRow(title) {
  return {
    title: title,
    showTitle: true,
    collapse: true,
    panels: []
  };
}

return function(callback) {
  var dashboard = {
    title: 'Alerts',
    time: {
      from: ARGS.from || 'now-6h',
      to: ARGS.to || 'now'
    },
    sharedCrosshair: true,
    hideControls: true,
    editable: false,
    rows: []
  };

  var datasourceName = ARGS.datasource;
  services.datasourceSrv.get(datasourceName)
  .then(function(datasource) {
    if (!datasource) {
      // not found
      callback({});
    }

    services.$q.when($.ajax({ method: 'GET', url: datasource.url + '/alerts', dataType: 'html', xhrFields: { withCredentials: true } }))
    .then(function(alertHTML) {
      $(alertHTML).find('.alert_header').each(function() {
        var alertName = $(this).find('b').text();
        var alertDetails = $(this).next();
        var alertRule = alertDetails.find('code').text();
        var alertExpression = alertDetails.find('a:last').text();
        var activeAlert = alertDetails.find('table').prop('outerHTML');

        var row = createRow(alertName.charAt(0) + alertName.slice(1).replace(/([A-Z])/g, ' $1'));

        var parsedExpression = alertExpression.match(/(.*) ([<>=]+ ([\d.]+))?$/);
        var graphPanel = createGraphPanel(alertName, datasourceName, [ { expr: parsedExpression[1], legendFormat: '{{instance}}' } ], 8);
        if (parsedExpression[3]) {
          var threshold = parseFloat(parsedExpression[3]);
          graphPanel.grid = {
            leftMax: threshold * 1.2,
            threshold1: threshold,
            threshold1Color: 'rgba(255, 0, 0, 255)',
            thresholdLine: true
          };
        }
        row.panels.push(graphPanel);
        var rulePanel = createTextPanel(alertName + ' Rule', '<pre>' + alertRule + '</pre>', 4);
        row.panels.push(rulePanel);

        if (activeAlert) {
          var activeAlertPanel = createTextPanel(alertName + ' Active', activeAlert, 12);
          row.panels.push(activeAlertPanel);
          row.collapse = false;
        }

        dashboard.rows.push(row);
      });

      callback(dashboard);
    });
  });
};
