import { describe, expect, it } from 'vitest';
import type { HttpClient } from '../../../../shared/api-client/http';
import type { TimerDTO } from '../../../../shared/api-client/generated/focus';
import { defaultFocusConfig as cfg } from '../../domain/focus';
import { FocusTimerApi } from './focusApi';

const dto = (over: Partial<TimerDTO> = {}): TimerDTO => ({
  phase: 'work',
  done: 0,
  running: false,
  remainingMs: 25 * 60_000,
  totalMs: 25 * 60_000,
  config: { work: 25, short: 5, long: 15, longEvery: 4 },
  ...over,
});

function fakeHttp(answer: TimerDTO, refusePut = false) {
  const calls: string[] = [];
  const http = {
    get: async () => (calls.push('GET'), answer),
    put: async (path: string) => {
      calls.push(`PUT ${path}`);
      if (refusePut) throw new Error('forbidden');
      return dto({ config: { work: 50, short: 10, long: 20, longEvery: 2 }, remainingMs: 50 * 60_000 });
    },
    post: async (path: string) => (calls.push(`POST ${path}`), dto({ running: true, remainingMs: 1000 })),
  } as unknown as HttpClient;
  return { http, calls };
}

describe('FocusTimerApi (#631)', () => {
  it('counts down from the time left on the browser clock', async () => {
    const { http } = fakeHttp(dto({ running: true, remainingMs: 5000 }));
    const s = await new FocusTimerApi(http, () => 100).get(cfg);
    expect(s.endsAt).toBe(5100);
  });

  it('stores the widget lengths once when they differ', async () => {
    const { http, calls } = fakeHttp(dto());
    const s = await new FocusTimerApi(http).get({ ...cfg, work: 50, short: 10, long: 20, longEvery: 2 });
    expect(calls).toEqual(['GET', 'PUT /api/focus/timer/config']);
    expect(s.left).toBe(50 * 60_000);
  });

  it('shows the timer as it is when the lengths may not be stored, and asks once', async () => {
    const { http, calls } = fakeHttp(dto(), true);
    const api = new FocusTimerApi(http);
    const other = { ...cfg, work: 50 };
    expect((await api.get(other)).left).toBe(25 * 60_000);
    await api.get(other);
    expect(calls).toEqual(['GET', 'PUT /api/focus/timer/config', 'GET']);
  });

  it('sends controls to the action path', async () => {
    const { http, calls } = fakeHttp(dto());
    await new FocusTimerApi(http, () => 0).control('pause');
    expect(calls).toEqual(['POST /api/focus/timer/pause']);
  });
});
