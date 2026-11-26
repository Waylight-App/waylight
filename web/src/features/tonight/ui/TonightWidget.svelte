<script lang="ts">
  import type { Snippet } from 'svelte';
  import { intlLocale, t } from '../../../shared/i18n';
  import {
    degrees,
    formatDate,
    formatTime,
    parseTonightConfig,
    sentenceCase,
    warningLevel,
    worstWarning,
  } from '../domain/tonight';
  import type { TonightUseCases } from '../usecase/tonightUseCases';
  import { useTonight } from './useTonight.svelte';
  import WidgetHeader from '../../../shared/ui/WidgetHeader.svelte';

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
  <WidgetHeader
    title={data.weather?.place ? t('tonight.title_in', { place: data.weather.place }) : t('tonight.title')}
    icon="sun"
  />
  <div class="head">
    <div class="clock">
      <time class="time" datetime={data.now.toISOString()}>{formatTime(data.now, cfg, intlLocale())}</time>
      <div class="muted line">{formatDate(data.now, cfg, intlLocale())}</div>
    </div>
    {#if data.weather}
      {@const w = data.weather}
      <div class="now">
        <div class="temp">{degrees(w.temperature)}</div>
        <div class="muted line">
          {w.condition}{w.feelsLike !== undefined
            ? ` · ${t('tonight.feels', { t: degrees(w.feelsLike) })}`
            : ''}
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
      <span>{sentenceCase(warning.title)}{warning.validity ? ` ${warning.validity}` : ''}</span>
      {#if data.warnings.length > 1}
        <span class="muted">{t('tonight.more_warnings', { n: data.warnings.length - 1 })}</span>
      {/if}
    </p>
  {/if}
  {#if data.weather?.days.length}
    <ul class="days" aria-label={t('tonight.forecast')}>
      {#each data.weather.days as d (d.date.getTime())}
        <li title={d.label}>
          <span class="muted day">{weekday(d.date)}<span class="day-icon">{@render icon(d.icon)}</span></span>
          <span class="range">{degrees(d.max)} / {degrees(d.min)}</span>
          <span class="muted rain">{Math.round(d.rain)} %</span>
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
    gap: var(--space-6);
    padding: var(--space-6);
    border-radius: var(--radius-md);
    /* The header and the text inside read the lit colours. */
    --color-text: var(--color-lit-text);
    --color-text-muted: var(--color-lit-muted);
    --color-text-subtle: var(--color-lit-muted);
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
  /* 32/40 at medium; a larger box may grow the two figures with it. */
  .time,
  .temp {
    display: block;
    font-family: var(--font-display);
    font-size: clamp(var(--fs-lg), min(5cqi, 10cqh), 5rem);
    font-weight: 500;
    font-variant-numeric: tabular-nums;
    line-height: 1.25;
  }
  .now {
    text-align: right;
  }
  .line,
  .note {
    margin: 0;
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
  }
  /* A chip in the lit chip colour: the lit text at a low share of the card. */
  .warning {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    align-self: flex-start;
    gap: var(--space-1) var(--space-2);
    min-height: 32px;
    padding: var(--space-1) var(--space-3);
    border-radius: 999px;
    background: color-mix(in srgb, var(--color-lit-accent) 14%, var(--color-lit));
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    font-weight: 500;
  }
  .dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-lit-warning);
  }
  [data-level='problem'] .dot {
    background: var(--color-lit-danger);
  }
  .days {
    list-style: none;
    margin: 0;
    padding: var(--space-4) 0 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: var(--space-2);
    border-top: 1px solid color-mix(in srgb, var(--color-lit-text) 14%, transparent);
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
  }
  .days li {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-1);
    min-width: 0;
  }
  .day {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
  }
  .day-icon {
    display: inline-flex;
  }
  /* Line icons are 16 px everywhere. */
  .day-icon :global(svg) {
    width: 16px;
    height: 16px;
  }
  .range {
    font-size: var(--fs-sm);
    line-height: var(--lh-sm);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
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
