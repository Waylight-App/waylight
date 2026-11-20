import { afterEach, describe, expect, it, vi } from 'vitest';
import { addLiveAlert, dismissLiveAlert, liveAlerts } from './liveAlerts.svelte';

const n = (title: string, severity: 'info' | 'warning' | 'critical' = 'info') => ({
  title,
  body: '',
  severity,
  type: 'service',
  at: '',
});

describe('live toasts (#583)', () => {
  afterEach(() => {
    liveAlerts.list = [];
    vi.useRealTimers();
  });

  it('keeps the newest four and dismisses on request', () => {
    for (const t of ['a', 'b', 'c', 'd', 'e']) addLiveAlert(n(t));
    expect(liveAlerts.list.map((a) => a.title)).toEqual(['b', 'c', 'd', 'e']);
    dismissLiveAlert(liveAlerts.list[0]!.key);
    expect(liveAlerts.list.map((a) => a.title)).toEqual(['c', 'd', 'e']);
  });

  it('lets info go by itself but keeps critical ones', () => {
    vi.useFakeTimers();
    addLiveAlert(n('info'));
    addLiveAlert(n('critical', 'critical'));
    vi.advanceTimersByTime(15_000);
    expect(liveAlerts.list.map((a) => a.title)).toEqual(['critical']);
  });
});
