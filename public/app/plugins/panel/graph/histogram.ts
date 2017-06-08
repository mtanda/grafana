import _ from 'lodash';

/**
 * Convert series into array of series values.
 * @param data Array of series
 */
export function getSeriesValues(data: any): number[] {
  let values = [];

  // Count histogam stats
  for (let i = 0; i < data.length; i++) {
    let series = data[i];
    for (let j = 0; j < series.data.length; j++) {
      if (series.data[j][1] !== null) {
        values.push(series.data[j][1]);
      }
    }
  }

  return values;
}

/**
 * Convert array of values into timeseries-like histogram:
 * [[val_1, count_1], [val_2, count_2], ..., [val_n, count_n]]
 * @param values
 * @param bucketSize
 */
export function convertValuesToHistogram(values: number[], bucketSize: number): any[] {
  let histogram = {};

  for (let i = 0; i < values.length; i++) {
    let bound = getBucketBound(values[i], bucketSize);
    if (histogram[bound]) {
      histogram[bound] = histogram[bound] + 1;
    } else {
      histogram[bound] = 1;
    }
  }

  let histogam_series = _.map(histogram, (count, bound) => {
    return [Number(bound), count];
  });

  // Sort by Y axis values
  return _.sortBy(histogam_series, point => point[0]);
}

/**
 * Convert series into array of histogram data.
 * @param data Array of series
 * @param bucketSize
 * @param stack
 */
export function convertToHistogramData(data: any, bucketSize: number, stack = false): any[] {
  let seriesValues = [];
  if (stack) {
    seriesValues = data.map((series) => {
      return { series: series, values: getSeriesValues([series]) };
    });
  } else {
    seriesValues = data.map((series, i) => {
      if (i === 0) {
        return { series: data[i], values: getSeriesValues(data) };
      } else {
        return { series: data[i], values: [] };
      }
    });
  }
  return seriesValues.map((sv) => {
    let series = sv.series;
    series.histogram = true;
    if (sv.values.length > 0) {
      let histogram = convertValuesToHistogram(sv.values, bucketSize);
      series.data = histogram;
    } else {
      series.data = [];
    }
    return series;
  });
}

function getBucketBound(value: number, bucketSize: number): number {
  return Math.floor(value / bucketSize) * bucketSize;
}
