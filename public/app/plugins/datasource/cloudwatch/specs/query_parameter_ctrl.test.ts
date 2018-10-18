import { CloudWatchQueryParameterCtrl } from '../query_parameter_ctrl';

describe('QueryParameterCtrl', () => {
  const ctx = {} as any;

  beforeEach(() => {
    ctx.ctrl = new CloudWatchQueryParameterCtrl({}, {}, {}, {}, {});
  });

  describe('should generate correct targetFull', () => {
    const targets = [
      {
        refId: 'A',
        id: 'id1',
        region: 'us-east-1',
        namespace: 'AWS/EC2',
        metricName: 'CPUUtilization',
        statistics: [
          'Average'
        ]
      },
      {
        refId: 'B',
        id: 'id2',
        expression: 'id1*2',
      },
      {
        refId: 'C',
        id: 'id3',
        expression: 'id2*2',
      },
      {
        refId: 'D',
        id: 'id4',
        region: 'us-west-1',
        namespace: 'AWS/EC2',
        metricName: 'CPUUtilization',
        statistics: [
          'Average'
        ]
      },
    ];
    let target: any = {
      refId: 'C',
      id: 'id3',
      expression: 'id2*2',
    };
    ctx.ctrl.renderTargetFull(target, targets)
    expect(target.targetFull.lenght).toBe(3);
    expect(target.targetFull[0].refId).toBe('A');
    expect(target.targetFull[0].region).toBe('ap-northeast-1');
    expect(target.targetFull[1].refId).toBe('B');
    expect(target.targetFull[1].region).toBe('ap-northeast-1');
    expect(target.targetFull[2].refId).toBe('C');
    expect(target.targetFull[2].region).toBe('ap-northeast-1');
  });
});
