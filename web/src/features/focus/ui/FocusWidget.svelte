<script lang="ts">
  import Icon from '../../../shared/ui/icons/Icon.svelte';
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
  import { hasHeaderMeta, useHeaderMeta } from '../../../shared/ui/widgetHeaderMeta';

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
  // The accent marks the best day of the week, the one bar worth a look (#815).
  const best = $derived(total > 0 ? week.findIndex((d) => d.sessions === max) : -1);
  // The week's total is the card's header meta (#809); without a card it stays under the bars.
  const inHeader = hasHeaderMeta();
  const setMeta = useHeaderMeta();
  $effect(() => setMeta(week.length > 0 ? { text: t('focus.week_total', { n: total }) } : null));
</script>

<!-- Time and controls on one line, the week's bars below (#818). -->
<div class="focus" class:running={isRunning(timer)} data-phase={timer.phase}>
  <div class="top">
    <div class="time">
      <p class="phase">
        {phaseLabel(timer.phase)}
        <span
          class="dots"
          role="img"
          aria-label={t('focus.cycle', { done: timer.done % cfg.longEvery, of: cfg.longEvery })}
        >
          {#each Array.from({ length: cfg.longEvery }, (_, i) => i) as i (i)}
            <span class:filled={i < timer.done % cfg.longEvery || (timer.phase === 'long' && timer.done > 0)}
            ></span>
          {/each}
        </span>
      </p>
      <p class="clock" role="timer" aria-live="off">{formatClock(left)}</p>
    </div>
    <div class="controls">
      {#if isRunning(timer)}
        <button class="btn btn-primary" type="button" onclick={() => act('pause')}>{t('focus.pause')}</button>
      {:else}
        <button class="btn btn-primary" type="button" onclick={() => act('start')}>{t('focus.start')}</button>
      {/if}
      <button class="btn" type="button" onclick={() => act('skip')}>{t('focus.skip')}</button>
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('focus.reset')}
        title={t('focus.reset')}
        onclick={() => act('reset')}><Icon name="history" size={16} /></button
      >
    </div>
  </div>
  {#if error}<p class="error">{error}</p>{/if}
  {#if week.length > 0}
    <figure class="week" aria-label={t('focus.week_total', { n: total })}>
      {#each week as d, i (d.date)}
        <div class="day" title={t('focus.day_title', { n: d.sessions, minutes: d.minutes })}>
          <span class="track"
            ><span class="bar" class:best={i === best} style:height="{(d.sessions / max) * 100}%"
            ></span></span
          >
          <span class="label">{weekday(d.date)}</span>
        </div>
      {/each}
      {#if !inHeader}<figcaption class="muted">{t('focus.week_total', { n: total })}</figcaption>{/if}
    </figure>
  {/if}
</div>

<style>
  .focus {
    display: grid;
    gap: var(--space-4);
  }
  p {
    margin: 0;
  }
  .top {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-4);
  }
  .phase {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    color: var(--color-text-subtle);
  }
  .clock {
    font-family: var(--font-display);
    font-size: var(--fs-lg);
    line-height: var(--lh-lg);
    font-weight: 500;
    font-variant-numeric: tabular-nums;
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
    gap: var(--space-2);
  }
  /* 64 px in all: 40 px of bars and the day letter; only the best day is amber. */
  .week {
    margin: 0;
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: var(--space-2);
  }
  .day {
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    align-items: center;
    gap: var(--space-1);
  }
  .track {
    display: flex;
    align-items: flex-end;
    justify-content: center;
    width: 100%;
    height: 40px;
  }
  .bar {
    width: 24px;
    max-width: 100%;
    min-height: 4px;
    background: var(--color-surface-2);
    border-radius: var(--radius-xs);
  }
  .bar.best {
    background: var(--color-accent);
  }
  .label {
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    color: var(--color-text-subtle);
  }
  figcaption {
    grid-column: 1 / -1;
    font-size: var(--fs-xs);
  }
</style>
