import { describe, expect, it, vi } from 'vitest';
import type { Timer } from '../domain/timer';
import { type TimerNotifier, Timers, type TimerScheduler, type TimerStore } from './timers';

function fakeClock(start: number) {
  let now = start;
  const pending: { at: number; fire: () => void; off: boolean }[] = [];
  const scheduler: TimerScheduler = {
    now: () => now,
    at(at, fire) {
      const p = { at, fire, off: false };
      pending.push(p);
      return () => (p.off = true);
    },
  };
  const advance = (ms: number) => {
    now += ms;
    pending.filter((p) => !p.off && p.at <= now).forEach((p) => ((p.off = true), p.fire()));
  };
  return { scheduler, advance };
}

const memoryStore = (initial: Timer[] = []): TimerStore & { saved: Timer[] } => {
  const s = {
    saved: initial,
    load: () => s.saved,
    save: (t: Timer[]) => (s.saved = t),
  };
  return s;
};

const notifier = (permission: NotificationPermission = 'granted'): TimerNotifier => ({
  permission: () => permission,
  request: vi.fn().mockResolvedValue(permission),
  show: vi.fn(),
});

describe('Timers (#474)', () => {
  it('fires a timer when it runs out and forgets it', () => {
    const { scheduler, advance } = fakeClock(1000);
    const store = memoryStore();
    const n = notifier();
    const timers = new Timers(store, scheduler, n);
    const fired = vi.fn();
    timers.onFire(fired);
    timers.start({ kind: 'timer', end: 61_000, label: 'Pizza', tomorrow: false });
    expect(store.saved).toHaveLength(1);
    advance(59_000);
    expect(fired).not.toHaveBeenCalled();
    advance(1000);
    expect(fired).toHaveBeenCalledOnce();
    expect(n.show).toHaveBeenCalledWith('Pizza', '');
    expect(timers.list()).toEqual([]);
    expect(store.saved).toEqual([]);
  });

  it('can be cancelled', () => {
    const { scheduler, advance } = fakeClock(0);
    const timers = new Timers(memoryStore(), scheduler, notifier());
    const fired = vi.fn();
    timers.onFire(fired);
    const t = timers.start({ kind: 'timer', end: 1000, label: '', tomorrow: false });
    timers.cancel(t.id);
    advance(2000);
    expect(fired).not.toHaveBeenCalled();
  });

  it('keeps timers across a reload and fires missed ones once', async () => {
    const store = memoryStore([
      { id: 'a', kind: 'timer', end: 5000, label: 'later', tomorrow: false },
      { id: 'b', kind: 'alarm', end: 500, label: 'missed', tomorrow: false },
    ]);
    const { scheduler } = fakeClock(1000);
    const timers = new Timers(store, scheduler, notifier());
    const fired = vi.fn();
    timers.onFire(fired);
    expect(timers.list().map((t) => t.id)).toEqual(['a']);
    await Promise.resolve();
    expect(fired).toHaveBeenCalledWith(expect.objectContaining({ id: 'b' }));
  });

  it('asks for notification permission on first use and still toasts without it', () => {
    const { scheduler, advance } = fakeClock(0);
    const n = notifier('default');
    const timers = new Timers(memoryStore(), scheduler, n);
    const fired = vi.fn();
    timers.onFire(fired);
    timers.start({ kind: 'timer', end: 10, label: '', tomorrow: false });
    expect(n.request).toHaveBeenCalled();
    advance(10);
    expect(fired).toHaveBeenCalled();
    expect(n.show).not.toHaveBeenCalled();
  });
});
