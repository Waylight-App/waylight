import {
  browserScheduler,
  chime,
  localTimerStore,
  timerNotifier,
} from '../features/search/adapter/browserTimers';
import type { Timer } from '../features/search/domain/timer';
import { Timers } from '../features/search/usecase/timers';
import { announce } from '../shared/a11y/live.svelte';
import { t } from '../shared/i18n';
import { timerAlerts } from './timerAlerts.svelte';

/** Title of a timer that ran out, as in the notification and the toast. */
export function timerTitle(timer: Timer): string {
  if (timer.label) return timer.label;
  return t(timer.kind === 'alarm' ? 'timers.alarm_done' : 'timers.timer_done');
}

/**
 * The timers of this browser (#474), one instance per page. Loaded with the
 * search panel, or by the shell after a reload when timers are stored.
 */
export const timers = new Timers(localTimerStore, browserScheduler, timerNotifier, timerTitle);

timers.onFire((timer) => {
  timerAlerts.list = [...timerAlerts.list, timer];
  announce(timerTitle(timer), { repeat: true });
  chime();
});
