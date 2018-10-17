import { CloudWatchQueryParameterCtrl } from '../query_parameter_ctrl';

describe('QueryParameterCtrl', () => {
  const ctx = {} as any;

  beforeEach(() => {

    ctx.ctrl = new CloudWatchQueryParameterCtrl({}, {}, {}, {}, {});
  });

  describe('generate targetFull', () => {
    console.log('here');
    console.log(ctx.ctrl.expressionChanged);
  });
});
