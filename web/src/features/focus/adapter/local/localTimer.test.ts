import { describe, expect, it } from 'vitest';
import { defaultFocusConfig as cfg } from '../../domain/focus';
import { LocalTimer, type TimerStorage } from './localTimer';

const MIN = 60_000;

function setup() {
  const data = new Map<string, string>();
  const storage: TimerStorage = {
    getItem: (k) => data.get(k) ?? null,
    setItem: (k, v) => void data.set(k, v),
  };
  const recorded: [number, Date][] = [];
  let now = 1_000_000;
  const timer = new LocalTimer(
    {
      record: async (minutes, at) => {
        recorded.push([minutes, at]);
        return [];
      },
    },
    () => storage,
    () => now,
  );
  return { timer, recorded, data, advance: (ms: number) => (now += ms), at: () => now };
}

describe('LocalTimer (#631)', () => {
  it('keeps the timer in storage across reads', async () => {
    const { timer, data, advance } = setup();
    const started = await timer.control('start', cfg);
    expect(started.endsAt).not.toBeNull();
    expect(data.size).toBe(1);
    advance(MIN);
    const paused = await timer.control('pause', cfg);
    expect(paused.left).toBe(24 * MIN);
    expect((await timer.get(cfg)).left).toBe(24 * MIN);
  });

  it('records a finished focus phase once and prepares the break', async () => {
    const { timer, recorded, advance, at } = setup();
    await timer.control('start', cfg);
    const end = at() + 25 * MIN;
    advance(60 * MIN);
    const next = await timer.get(cfg);
    await timer.get(cfg);
    expect(next.phase).toBe('short');
    expect(next.endsAt).toBeNull();
    expect(recorded).toEqual([[25, new Date(end)]]);
  });

  it('works without storage', async () => {
    const timer = new LocalTimer({ record: async () => [] }, () => undefined);
    expect((await timer.control('start', cfg)).endsAt).not.toBeNull();
  });
});
