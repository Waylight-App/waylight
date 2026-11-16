import type { Timer } from '../domain/timer';
import type { TimerNotifier, TimerScheduler, TimerStore } from '../usecase/timers';

const KEY = 'trisse.timers';

/** Running timers in localStorage, so they survive a reload (#474). */
export const localTimerStore: TimerStore = {
  load() {
    try {
      const raw = JSON.parse(localStorage.getItem(KEY) ?? '[]') as unknown;
      return Array.isArray(raw) ? raw.filter(isTimer) : [];
    } catch {
      return [];
    }
  },
  save(timers) {
    try {
      localStorage.setItem(KEY, JSON.stringify(timers));
    } catch {
      // Private mode or a full storage: the timer still runs in this tab.
    }
  },
};

function isTimer(x: unknown): x is Timer {
  const t = x as Timer;
  return (
    typeof t === 'object' &&
    t !== null &&
    typeof t.id === 'string' &&
    typeof t.end === 'number' &&
    (t.kind === 'timer' || t.kind === 'alarm') &&
    typeof t.label === 'string'
  );
}

/**
 * setTimeout caps delays at about 24.8 days and drifts in background tabs,
 * so long waits are split and the end time is checked again on wake-up.
 */
const MAX_STEP = 60_000;

export const browserScheduler: TimerScheduler = {
  now: () => Date.now(),
  at(at, fire) {
    let handle: ReturnType<typeof setTimeout>;
    const step = () => {
      const left = at - Date.now();
      if (left <= 0) return fire();
      handle = setTimeout(step, Math.min(left, MAX_STEP));
    };
    step();
    return () => clearTimeout(handle);
  },
};

/** Browser notifications; the same API the focus timer uses. */
export const timerNotifier: TimerNotifier = {
  permission: () => (typeof Notification === 'undefined' ? 'unsupported' : Notification.permission),
  request: async () =>
    typeof Notification === 'undefined' ? 'unsupported' : Notification.requestPermission(),
  show: (title, body) => {
    try {
      new Notification(title, { body, tag: 'trisse-timer' });
    } catch {
      // Some browsers only allow notifications from a service worker.
    }
  },
};

/** A short two-tone chime; silent when audio is blocked. */
export function chime(): void {
  try {
    const ctx = new AudioContext();
    [880, 660].forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.frequency.value = freq;
      gain.gain.setValueAtTime(0.15, ctx.currentTime + i * 0.25);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + i * 0.25 + 0.22);
      osc.connect(gain).connect(ctx.destination);
      osc.start(ctx.currentTime + i * 0.25);
      osc.stop(ctx.currentTime + i * 0.25 + 0.25);
    });
    setTimeout(() => void ctx.close(), 800);
  } catch {
    // No audio: the toast and the notification still show.
  }
}
