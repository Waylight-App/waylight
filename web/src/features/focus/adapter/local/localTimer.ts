import {
  applyConfig,
  control,
  type FocusConfig,
  initialState,
  parseState,
  tick,
  type TimerAction,
  type TimerState,
} from '../../domain/focus';
import type { FocusGateway, TimerGateway } from '../../usecase/focusUseCases';

/** The part of Web Storage the timer needs, so tests can pass a map. */
export type TimerStorage = Pick<Storage, 'getItem' | 'setItem'>;

const STORAGE_KEY = 'waylight-focus-timer';

/**
 * TimerGateway of local mode (ADR 0019): the timer lives in the browser's
 * storage and finished focus phases go to the local session store. It
 * follows the same rules as the server's timer (#631).
 */
export class LocalTimer implements TimerGateway {
  constructor(
    private readonly sessions: Pick<FocusGateway, 'record'>,
    private readonly storage: () => TimerStorage | undefined = () => globalThis.localStorage,
    private readonly now: () => number = Date.now,
  ) {}

  async get(cfg: FocusConfig): Promise<TimerState> {
    const stored = this.read(cfg);
    const r = tick(applyConfig(stored, cfg), cfg, this.now());
    if (r.state !== stored) this.write(r.state);
    if (r.finished?.phase === 'work') await this.sessions.record(cfg.work, new Date(r.finished.at));
    return r.state;
  }

  async control(action: TimerAction, cfg: FocusConfig): Promise<TimerState> {
    const next = control(await this.get(cfg), action, cfg, this.now());
    this.write(next);
    return next;
  }

  private read(cfg: FocusConfig): TimerState {
    try {
      const raw = this.storage()?.getItem(STORAGE_KEY);
      return raw ? parseState(JSON.parse(raw), cfg) : initialState(cfg);
    } catch {
      return initialState(cfg);
    }
  }

  private write(s: TimerState): void {
    try {
      this.storage()?.setItem(STORAGE_KEY, JSON.stringify(s));
    } catch {
      // Private mode: the timer still works until the page is closed.
    }
  }
}
