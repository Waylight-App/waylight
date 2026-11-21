import type { HttpClient } from '../../../../shared/api-client/http';
import type { TimerDTO, WeekDTO } from '../../../../shared/api-client/generated/focus';
import type {
  SessionInputDTO,
  TimerConfigInputDTO,
} from '../../../../shared/api-client/generated/focus-requests';
import {
  type FocusConfig,
  type FocusDay,
  fromSnapshot,
  type Phase,
  sameLengths,
  type TimerAction,
  type TimerSnapshot,
  type TimerState,
} from '../../domain/focus';
import type { FocusGateway, Notifier, TimerGateway } from '../../usecase/focusUseCases';

const toDays = (d: WeekDTO): FocusDay[] =>
  d.days.map((x) => ({ date: x.date, sessions: x.sessions, minutes: x.minutes }));

/** FocusGateway backed by the REST API. */
export class FocusApi implements FocusGateway {
  constructor(private readonly http: HttpClient) {}

  async stats(): Promise<FocusDay[]> {
    return toDays(await this.http.get<WeekDTO>('/api/focus/stats'));
  }

  async record(minutes: number, endedAt: Date): Promise<FocusDay[]> {
    const body: SessionInputDTO = { minutes, endedAt: endedAt.toISOString() };
    return toDays(await this.http.post<WeekDTO>('/api/focus/sessions', body));
  }
}

const PHASES: readonly Phase[] = ['work', 'short', 'long'];

/** Maps GET /api/focus/timer; an unknown phase is read as focus. */
export function toSnapshot(d: TimerDTO): TimerSnapshot {
  return {
    phase: PHASES.find((p) => p === d.phase) ?? 'work',
    done: d.done,
    running: d.running,
    remainingMs: d.remainingMs,
    totalMs: d.totalMs,
    config: d.config,
  };
}

/**
 * TimerGateway backed by the account's timer on the server (#631). The
 * widget's phase lengths are stored when they differ from the server's;
 * a viewer who may not store them (a guest) sees the timer as it is.
 */
export class FocusTimerApi implements TimerGateway {
  /** Lengths the server refused, so a poll does not ask again. */
  private refused: string | null = null;

  constructor(
    private readonly http: HttpClient,
    private readonly now: () => number = Date.now,
  ) {}

  async get(cfg: FocusConfig): Promise<TimerState> {
    const snapshot = toSnapshot(await this.http.get<TimerDTO>('/api/focus/timer'));
    const key = JSON.stringify([cfg.work, cfg.short, cfg.long, cfg.longEvery]);
    if (sameLengths(snapshot.config, cfg) || this.refused === key) return fromSnapshot(snapshot, this.now());
    const body: TimerConfigInputDTO = {
      work: cfg.work,
      short: cfg.short,
      long: cfg.long,
      longEvery: cfg.longEvery,
    };
    try {
      return this.state(await this.http.put<TimerDTO>('/api/focus/timer/config', body));
    } catch {
      this.refused = key;
      return fromSnapshot(snapshot, this.now());
    }
  }

  async control(action: TimerAction): Promise<TimerState> {
    return this.state(await this.http.post<TimerDTO>(`/api/focus/timer/${action}`));
  }

  private state(d: TimerDTO): TimerState {
    return fromSnapshot(toSnapshot(d), this.now());
  }
}

/** Notifier backed by the browser's Notification API. */
export const browserNotifier: Notifier = {
  permission: () => (typeof Notification === 'undefined' ? 'unsupported' : Notification.permission),
  request: async () =>
    typeof Notification === 'undefined' ? 'unsupported' : Notification.requestPermission(),
  show: (title, body) => {
    try {
      new Notification(title, { body, tag: 'waylight-focus' });
    } catch {
      // Some browsers only allow notifications from a service worker.
    }
  },
};
