<script lang="ts">
  import type { Snippet } from 'svelte';
  import { intlLocale, t } from '../../../shared/i18n';
  import {
    degrees,
    formatDate,
    formatTime,
    parseTonightConfig,
    warningLevel,
    worstWarning,
  } from '../domain/tonight';
  import type { TonightUseCases } from '../usecase/tonightUseCases';
  import { useTonight } from './useTonight.svelte';

  /** Props of the Tonight card: the clock and the current weather in one lit card. */
  interface Props {
    tonight: TonightUseCases;
    config: Record<string, unknown>;
    /** Draws the icon of a weather kind; the weather feature owns the icons. */
    icon: Snippet<[string]>;
  }
  let { tonight, config, icon }: Props = $props();

  const cfg = $derived(parseTonightConfig(config));
  const data = useTonight(() => ({ tonight, config }));
  const warning = $derived(worstWarning(data.warnings));
  const weekday = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', timeZone: cfg.timeZone || undefined });
</script>

<section class="lit" aria-label={t('tonight.label')}>
  <div class="head">
    <div class="clock">
      <time class="time" datetime={data.now.toISOString()}>{formatTime(data.now, cfg, intlLocale())}</time>
      <div class="date">{formatDate(data.now, cfg, intlLocale())}</div>
    </div>
    {#if data.weather}
      {@const w = data.weather}
      <div class="now">
        <span class="icon">{@render icon(w.icon)}</span>
        <div>
          <div class="temp">{degrees(w.temperature)}</div>
          <div class="muted">{w.condition}{w.place ? ` · ${w.place}` : ''}</div>
        </div>
      </div>
    {:else if data.loaded}
      <p class="muted note">{t('tonight.choose_a_location')}</p>
    {:else if data.error}
      <p class="muted note">{t('tonight.weather_unavailable')}</p>
    {/if}
  </div>
  {#if warning}
    <p class="warning" data-level={warningLevel(warning)}>
      <span class="dot" aria-hidden="true"></span>
      <span class="sr-only">{t('tonight.warning')}:</span>
      <strong>{warning.title}</strong>
      {#if warning.validity}<span class="muted">{warning.validity}</span>{/if}
      {#if data.warnings.length > 1}
        <span class="muted">{t('tonight.more_warnings', { n: data.warnings.length - 1 })}</span>
      {/if}
    </p>
  {/if}
  {#if data.weather?.days.length}
    <ul class="days" aria-label={t('tonight.forecast')}>
      {#each data.weather.days as d (d.date.getTime())}
        <li title={d.label}>
          <span class="muted">{weekday(d.date)}</span>
          <span class="day-icon">{@render icon(d.icon)}</span>
          <span>{degrees(d.max)} <span class="muted">{degrees(d.min)}</span></span>
          {#if d.rain > 0}<span class="rain">{Math.round(d.rain)}%</span>{/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  /* The one lit card of the screen: it inverts the scheme (design guide, Surfaces). */
  .lit {
    box-sizing: border-box;
    min-height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-4) var(--space-6);
    border-radius: var(--radius-md);
    background: var(--color-lit);
    color: var(--color-lit-text);
    box-shadow: var(--shadow);
    overflow: hidden;
  }
  .muted {
    color: var(--color-lit-muted);
  }
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-3) var(--space-4);
  }
  .time {
    display: block;
    font-family: var(--font-display);
    font-size: clamp(2rem, min(12cqi, 22cqh), 5rem);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 1.05;
  }
  .date {
    font-size: var(--fs-sm);
    color: var(--color-lit-muted);
  }
  .now {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .icon {
    display: flex;
    font-size: clamp(2rem, min(11cqi, 18cqh), 4rem);
    line-height: 1;
    color: var(--color-lit-accent);
  }
  .temp {
    font-family: var(--font-display);
    font-size: clamp(1.6rem, min(8cqi, 14cqh), 3rem);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    line-height: 1;
  }
  .note {
    margin: 0;
    font-size: var(--fs-sm);
  }
  .warning {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: var(--space-1) var(--space-2);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-sm);
    background: color-mix(in srgb, var(--color-lit-text) 9%, transparent);
    font-size: var(--fs-xs);
  }
  .dot {
    flex: none;
    width: 0.6em;
    height: 0.6em;
    border-radius: 50%;
    align-self: center;
    background: var(--color-lit-warning);
  }
  [data-level='problem'] .dot {
    background: var(--color-lit-danger);
  }
  .days {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: var(--space-2);
    text-align: center;
    font-size: var(--fs-sm);
  }
  .days li {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-1);
    border-radius: var(--radius-sm);
    background: color-mix(in srgb, var(--color-lit-text) 9%, transparent);
  }
  .days .muted,
  .rain {
    font-size: var(--fs-xs);
  }
  .day-icon {
    font-size: clamp(1.2rem, 6cqi, 2.2rem);
    color: var(--color-lit-accent);
  }
  .rain {
    color: var(--color-lit-accent);
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
  /* Narrow boxes keep the forecast to the days that fit. */
  @container widget (max-width: 259px) {
    .days {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .days li:nth-child(n + 3) {
      display: none;
    }
  }
  /* Medium tiles are short: the rain chance goes before anything else. */
  @container widget (max-height: 300px) {
    .rain {
      display: none;
    }
  }
  /* Short boxes drop the forecast first. */
  @container widget (max-height: 190px) and (min-width: 260px) {
    .days {
      display: none;
    }
  }
</style>
