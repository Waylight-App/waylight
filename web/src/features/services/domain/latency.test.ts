import { describe, expect, it } from 'vitest';
import { latencyShapes, latencySummary } from './latency';

const at = new Date(0);
const up = (latencyMs: number) => ({ at, latencyMs, online: true });
const down = (latencyMs = 10_000) => ({ at, latencyMs, online: false });

describe('latency chart (#862)', () => {
  it('sums up answered checks only and counts outages as runs', () => {
    expect(latencySummary([up(10), down(), down(), up(30), down(), up(20)])).toEqual({
      latest: 20,
      average: 20,
      max: 30,
      outages: 2,
      down: false,
    });
    expect(latencySummary([down()])).toEqual({
      latest: null,
      average: null,
      max: null,
      outages: 1,
      down: true,
    });
    expect(latencySummary([])).toEqual({ latest: null, average: null, max: null, outages: 0, down: false });
  });

  it('breaks the line at outages and scales by answered checks', () => {
    const s = latencyShapes([up(50), up(100), down(), up(100)], 30, 20);
    expect(s.lines).toEqual(['0,10 10,0', '30,0']);
    expect(s.outages).toEqual([{ x: 15, width: 10 }]);
  });

  it('keeps outage bars inside the box', () => {
    expect(latencyShapes([down(), up(1)], 10, 20).outages).toEqual([{ x: 0, width: 5 }]);
  });
});
