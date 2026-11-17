import { running, type Timer, type TimerRequest } from '../domain/timer';

/** Output port: where running timers are kept (the browser, so they survive a reload). */
export interface TimerStore {
  load(): Timer[];
  save(timers: Timer[]): void;
}

/** Output port: browser notifications, as in the focus timer. */
export interface TimerNotifier {
  permission(): NotificationPermission | 'unsupported';
  request(): Promise<NotificationPermission | 'unsupported'>;
  show(title: string, body: string): void;
}

/**
 * Output port: the clock that wakes a timer (#474). The browser adapter uses
 * setTimeout; tests use a fake one.
 */
export interface TimerScheduler {
  now(): number;
  /** Calls `fire` at `at`; returns a function that cancels it. */
  at(at: number, fire: () => void): () => void;
}

/**
 * Timers and alarms of the search field (#474). Running timers are stored
 * with their end time, scheduled again after a reload, and fire the
 * listener (the in-app toast) and a browser notification when they run out.
 * They only fire while a waylight tab or the installed app is open.
 */
export class Timers {
  private timers: Timer[];
  private readonly cancels = new Map<string, () => void>();
  private readonly listeners = new Set<(t: Timer) => void>();
  private readonly changes = new Set<(list: Timer[]) => void>();

  constructor(
    private readonly store: TimerStore,
    private readonly scheduler: TimerScheduler,
    readonly notifier: TimerNotifier,
    private readonly title: (t: Timer) => string = (t) => t.label || t.kind,
  ) {
    const now = scheduler.now();
    const stored = store.load();
    // Timers that ran out while no tab was open fire once on load.
    const missed = stored.filter((t) => t.end <= now);
    this.timers = running(stored, now);
    if (missed.length) this.persist();
    this.timers.forEach((t) => this.arm(t));
    queueMicrotask(() => missed.forEach((t) => this.fire(t, false)));
  }

  /** The running timers, soonest first. */
  list(): Timer[] {
    return [...this.timers];
  }

  /** Starts a timer; asks for notification permission on first use. */
  start(req: TimerRequest): Timer {
    const timer: Timer = { ...req, id: `${req.end}-${Math.random().toString(36).slice(2, 8)}` };
    this.timers = running([...this.timers, timer], this.scheduler.now());
    this.persist();
    this.arm(timer);
    if (this.notifier.permission() === 'default') void this.notifier.request();
    return timer;
  }

  cancel(id: string): void {
    this.cancels.get(id)?.();
    this.cancels.delete(id);
    this.timers = this.timers.filter((t) => t.id !== id);
    this.persist();
  }

  /** Called when a timer runs out (the in-app toast). Returns an unsubscribe function. */
  onFire(listener: (t: Timer) => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  /** Called whenever the list changes. Returns an unsubscribe function. */
  onChange(listener: (list: Timer[]) => void): () => void {
    this.changes.add(listener);
    return () => this.changes.delete(listener);
  }

  private arm(t: Timer) {
    this.cancels.set(
      t.id,
      this.scheduler.at(t.end, () => this.fire(t, true)),
    );
  }

  private fire(t: Timer, live: boolean) {
    this.cancels.delete(t.id);
    this.timers = this.timers.filter((x) => x.id !== t.id);
    if (live) this.persist();
    if (this.notifier.permission() === 'granted') this.notifier.show(this.title(t), '');
    this.listeners.forEach((l) => l(t));
  }

  private persist() {
    this.store.save(this.timers);
    const list = this.list();
    this.changes.forEach((l) => l(list));
  }
}
