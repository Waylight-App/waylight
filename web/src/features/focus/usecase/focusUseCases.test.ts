import { describe, expect, it } from 'vitest';
import { defaultFocusConfig, type FocusDay, initialState } from '../domain/focus';
import { type FocusGateway, FocusUseCases, type Notifier, type TimerGateway } from './focusUseCases';

class FakeGateway implements FocusGateway {
  recorded: [number, Date][] = [];
  async stats(): Promise<FocusDay[]> {
    return [];
  }
  async record(minutes: number, endedAt: Date): Promise<FocusDay[]> {
    this.recorded.push([minutes, endedAt]);
    return [{ date: '2026-11-12', sessions: 1, minutes }];
  }
}

class FakeNotifier implements Notifier {
  shown: string[] = [];
  constructor(private readonly perm: NotificationPermission | 'unsupported') {}
  permission() {
    return this.perm;
  }
  async request() {
    return this.perm;
  }
  show(title: string) {
    this.shown.push(title);
  }
}

const timers: TimerGateway & { calls: string[] } = {
  calls: [],
  async get(cfg) {
    this.calls.push('get');
    return initialState(cfg);
  },
  async control(action, cfg) {
    this.calls.push(action);
    return initialState(cfg);
  },
};

describe('FocusUseCases', () => {
  it('reads and controls the timer through its gateway', async () => {
    const uc = new FocusUseCases(new FakeGateway(), new FakeNotifier('default'), timers);
    await uc.timer(defaultFocusConfig);
    await uc.control('start', defaultFocusConfig);
    expect(timers.calls).toEqual(['get', 'start']);
  });

  it('notifies only when enabled and allowed', () => {
    const granted = new FakeNotifier('granted');
    const uc = new FocusUseCases(new FakeGateway(), granted, timers);
    uc.notify(false, 'a', '');
    uc.notify(true, 'b', '');
    const denied = new FakeNotifier('denied');
    new FocusUseCases(new FakeGateway(), denied, timers).notify(true, 'c', '');
    expect(granted.shown).toEqual(['b']);
    expect(denied.shown).toEqual([]);
  });
});
