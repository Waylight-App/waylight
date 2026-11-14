<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    bars,
    byDay,
    currentHour,
    type ElectricityPrices,
    formatCent,
    parseConfig,
  } from '../domain/electricity';
  import type { ElectricityUseCases } from '../usecase/electricityUseCases';

  /** Props of the electricity price widget (#288). */
  interface Props {
    electricity: ElectricityUseCases;
    config: Record<string, unknown>;
  }
  let { electricity, config }: Props = $props();
  const poll = usePolling();
  const POLL_MS = 15 * 60_000;

  const cfg = $derived(parseConfig(config));
  let prices = $state<ElectricityPrices | null>(null);
  let error = $state('');
  let now = $state(new Date());

  $effect(() => {
    const zone = cfg.zone;
    return poll(
      () =>
        electricity
          .prices(zone)
          .then((p) => ((prices = p), (error = ''), (now = new Date())))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      POLL_MS,
    );
  });

  const days = $derived(prices ? byDay(prices.hours) : []);
  const current = $derived(prices ? currentHour(prices.hours, now) : undefined);
  const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
</script>

{#if prices}
  <div class="head">
    {#if current}
      <span class="now"
        ><span class="muted">{t('electricity.now')}</span>
        <strong>{formatCent(current.centPerKWh, intlLocale())}</strong></span
      >
    {/if}
    <span class="legend"><span class="swatch cheap"></span> {t('electricity.cheapest')}</span>
  </div>
  {#each days as d, i (d.day)}
    {@const b = bars(d.hours)}
    <h4>{i === 0 ? t('electricity.today') : t('electricity.tomorrow')}</h4>
    <ul
      class="chart"
      style:--zero="{b.zero}%"
      aria-label={i === 0 ? t('electricity.today') : t('electricity.tomorrow')}
    >
      {#each d.hours as h, j (h.start.getTime())}
        {@const label = t('electricity.hour', {
          time: time(h.start),
          price: formatCent(h.centPerKWh, intlLocale()),
        })}
        <li
          class="bar"
          class:cheap={h.cheapest}
          class:negative={h.centPerKWh < 0}
          class:current={current === h}
          style:--h="{b.heights[j] ?? 0}%"
          title={label}
          aria-label={label}
        ></li>
      {/each}
    </ul>
  {/each}
  {#if days.length === 1}<p class="muted small">{t('electricity.tomorrow_later')}</p>{/if}
  <p class="muted small">
    {#if prices.stale}<span class="warn">{t('electricity.stale')}</span>{/if}
    {t('electricity.market_note')}
  </p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .now strong {
    font-size: 1.3rem;
  }
  .legend {
    font-size: max(0.8rem, 0.8em);
    color: var(--color-text-muted);
  }
  .swatch {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 2px;
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: max(0.8rem, 0.8em);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  /* One bar per hour; the zero line sits at --zero from the top. */
  .chart {
    position: relative;
    display: flex;
    gap: 2px;
    height: 72px;
    margin: 0;
    padding: 0;
    list-style: none;
    border-bottom: 1px solid var(--color-border);
  }
  .bar {
    position: relative;
    flex: 1;
  }
  .bar::before {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: calc(100% - var(--zero));
    height: var(--h);
    background: color-mix(in srgb, var(--color-accent) 45%, transparent);
    border-radius: 2px 2px 0 0;
  }
  .bar.negative::before {
    bottom: auto;
    top: var(--zero);
    border-radius: 0 0 2px 2px;
  }
  .bar.cheap::before,
  .swatch.cheap {
    background: var(--color-success);
  }
  .bar.current::before {
    outline: 2px solid var(--color-text);
    outline-offset: 1px;
  }
  .small {
    font-size: max(0.8rem, 0.85em);
    margin: var(--space-1) 0 0;
  }
  .warn {
    color: var(--color-warning);
  }
</style>
