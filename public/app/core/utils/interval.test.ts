import {
  populateInterval,
} from './interval';
import { DataQueryRequest, IntervalValues, TimeRange } from '@grafana/data';

describe('interval', () => {
  describe('populateInterval', () => {
    it('returns', () => {
      expect(populateInterval({}, {}, 100, '100')).toMatchObject({
      });
    });
  });
});
