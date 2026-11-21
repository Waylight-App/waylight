import type { FocusConfig, FocusDay, TimerAction, TimerState } from '../domain/focus';

/** Output port: focus backend. */
export interface FocusGateway {
  stats(): Promise<FocusDay[]>;
  /** Records a finished session and returns the updated statistics. */
  record(minutes: number, endedAt: Date): Promise<FocusDay[]>;
}

/**
 * Output port: the running timer (#631). On a server it is kept per account,
 * so the widget, its compact view and other devices share it; in local mode
 * it lives in the browser. Reading it advances a phase that ran out, and a
 * finished focus phase is recorded once.
 */
export interface TimerGateway {
  get(cfg: FocusConfig): Promise<TimerState>;
  control(action: TimerAction, cfg: FocusConfig): Promise<TimerState>;
}

/** Output port: browser notifications. */
export interface Notifier {
  permission(): NotificationPermission | 'unsupported';
  request(): Promise<NotificationPermission | 'unsupported'>;
  show(title: string, body: string): void;
}

/** Application logic of the focus timer (#58). */
export class FocusUseCases {
  constructor(
    private readonly gateway: FocusGateway,
    readonly notifier: Notifier,
    private readonly timers: TimerGateway,
  ) {}

  /** The timer with the widget's phase lengths. */
  timer(cfg: FocusConfig): Promise<TimerState> {
    return this.timers.get(cfg);
  }

  /** Starts, pauses, resets or skips the timer. */
  control(action: TimerAction, cfg: FocusConfig): Promise<TimerState> {
    return this.timers.control(action, cfg);
  }

  stats(): Promise<FocusDay[]> {
    return this.gateway.stats();
  }

  /** Notifies about the end of a phase if enabled and allowed. */
  notify(enabled: boolean, title: string, body: string): void {
    if (enabled && this.notifier.permission() === 'granted') this.notifier.show(title, body);
  }
}
