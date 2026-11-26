<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import {
    control,
    type FocusDay,
    formatClock,
    initialState,
    isRunning,
    parseFocusConfig,
    type Phase,
    remaining,
    type TimerAction,
    type TimerState,
  } from '../domain/focus';
  import type { FocusUseCases } from '../usecase/focusUseCases';

  /** Props of the focus timer widget (#58). */
  interface Props {
    focus: FocusUseCases;
    config: Record<string, unknown>;
  }
  let { focus, config }: Props = $props();

  /**
   * How often the timer is read again, so a start or pause from the compact
   * view or another device shows up here (#631).
   */
  const SYNC_MS = 15_000;
  const cfg = $derived(parseFocusConfig(config));

  // svelte-ignore state_referenced_locally
  let timer = $state<TimerState>(initialState(parseFocusConfig(config)));
  let now = $state(Date.now());
  let week = $state<FocusDay[]>([]);
  let error = $state('');

  const phaseLabel = (p: Phase) => t(`focus.phase.${p}`);

  async function finished(phase: Phase) {
    focus.notify(cfg.notify, t('focus.notify_title', { phase: phaseLabel(phase) }), t('focus.notify_body'));
    if (phase !== 'work') return;
    // The finished session was recorded where the timer lives.
    try {
      week = await focus.stats();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  /**
   * Takes a state read from the timer. When the shown phase had run out and
   * the timer moved on, the end is announced, whichever read noticed it.
   */
  function adopt(next: TimerState) {
    const before = timer;
    timer = next;
    error = '';
    if (
      isRunning(before) &&
      remaining(before, Date.now()) === 0 &&
      !isRunning(next) &&
      next.phase !== before.phase
    ) {
      void finished(before.phase);
    }
  }

  async function load() {
    try {
      adopt(await focus.timer(cfg));
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  // Reading cfg inside load makes changed settings read the timer again.
  const poll = usePolling();
  $effect(() => poll(load, SYNC_MS));

  async function act(action: TimerAction) {
    // Shown at once; the answer replaces it.
    timer = control(timer, action, cfg, Date.now());
    try {
      adopt(await focus.control(action, cfg));
    } catch (err) {
      error = errorMessage(err);
      void load().catch(() => undefined);
    }
  }

  let ending = false;
  /** Counts down and reads the timer once the phase has run out. */
  function step() {
    now = Date.now();
    if (ending || !isRunning(timer) || remaining(timer, now) > 0) return;
    ending = true;
    load()
      .catch(() => undefined)
      .finally(() => (ending = false));
  }

  // A one-second tick only while running; the end is computed from the
  // end time, so a throttled background tab still ends correctly.
  $effect(() => {
    if (!isRunning(timer)) return;
    step();
    const id = setInterval(step, 1000);
    document.addEventListener('visibilitychange', step);
    return () => {
      clearInterval(id);
      document.removeEventListener('visibilitychange', step);
    };
  });

  $effect(() => {
    focus
      .stats()
      .then((d) => (week = d))
      .catch((err) => (error = errorMessage(err)));
  });

  const left = $derived(remaining(timer, now));
  const max = $derived(Math.max(1, ...week.map((d) => d.sessions)));
  const weekday = (date: string) =>
    new Date(`${date}T12:00:00`).toLocaleDateString(intlLocale(), { weekday: 'narrow' });
  const total = $derived(week.reduce((n, d) => n + d.sessions, 0));
</script>

<div class="focus" class:running={isRunning(timer)} data-phase={timer.phase}>
  <p class="phase">{phaseLabel(timer.phase)}</p>
  <p class="clock" role="timer" aria-live="off">{formatClock(left)}</p>
  <p class="dots" role="img" aria-label={t('focus.cycle', { done: timer.done % cfg.longEvery, of: cfg.longEvery })}>
    {#each Array.from({ length: cfg.longEvery }, (_, i) => i) as i (i)}
      <span class:filled={i < timer.done % cfg.longEvery || (timer.phase === 'long' && timer.done > 0)}
      ></span>
    {/each}
  </p>
  <div class="controls">
    {#if isRunning(timer)}
      <button class="btn btn-primary" type="button" onclick={() => act('pause')}>{t('focus.pause')}</button>
    {:else}
      <button class="btn btn-primary" type="button" onclick={() => act('start')}>{t('focus.start')}</button>
    {/if}
    <button class="btn" type="button" onclick={() => act('reset')}>{t('focus.reset')}</button>
    <button class="btn" type="button" onclick={() => act('skip')}>{t('focus.skip')}</button>
  </div>
  {#if error}<p class="error">{error}</p>{/if}
  {#if week.length > 0}
    <figure class="week" aria-label={t('focus.week_total', { n: total })}>
      {#each week as d (d.date)}
        <div class="day" title={t('focus.day_title', { n: d.sessions, minutes: d.minutes })}>
          <span class="bar" style:height="{(d.sessions / max) * 100}%"></span>
          <span class="label">{weekday(d.date)}</span>
        </div>
      {/each}
      <figcaption class="muted">{t('focus.week_total', { n: total })}</figcaption>
    </figure>
  {/if}
</div>

<style>
  .focus {
    display: grid;
    justify-items: center;
    gap: var(--space-1);
    text-align: center;
  }
  p {
    margin: 0;
  }
  .phase {
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  .clock {
    font-family: var(--font-display);
    font-size: var(--fs-lg);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 1.1;
  }
  .running .clock {
    color: var(--color-accent-text);
  }
  [data-phase='short'] .clock,
  [data-phase='long'] .clock {
    color: var(--color-success, var(--color-accent));
  }
  .dots {
    display: flex;
    gap: var(--space-1);
  }
  .dots span {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    border: 1px solid var(--color-text-subtle);
  }
  .dots span.filled {
    background: var(--color-accent);
    border-color: var(--color-accent);
  }
  .controls {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-1);
  }
  .week {
    margin: var(--space-2) 0 0;
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: var(--space-1);
    width: 100%;
    max-width: 16rem;
  }
  .day {
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    align-items: center;
    height: 3rem;
  }
  .bar {
    width: 60%;
    min-height: 2px;
    background: var(--color-accent);
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  }
  .label {
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  figcaption {
    grid-column: 1 / -1;
    font-size: var(--fs-xs);
  }
</style>
