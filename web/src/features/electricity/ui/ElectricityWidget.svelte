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
  const DAY_LABEL = { today: 'electricity.today', tomorrow: 'electricity.tomorrow' } as const;

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

  const days = $derived(prices ? byDay(prices.hours, now) : []);
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
  {#each days as d (d.day)}
    {@const b = bars(d.hours)}
    <h4>{t(DAY_LABEL[d.when])}</h4>
    <ul class="chart" style:--zero="{b.zero}%" aria-label={t(DAY_LABEL[d.when])}>
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
  {#if days.length === 1 && days[0]?.when === 'today'}<p class="muted small">
      {t('electricity.tomorrow_later')}
    </p>{/if}
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
    font-size: var(--fs-md);
  }
  .legend {
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  .swatch {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: var(--radius-sm);
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  /* One bar per hour; the zero line sits at --zero from the top. */
  .chart {
    position: relative;
    display: flex;
    gap: var(--space-1);
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
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  }
  .bar.negative::before {
    bottom: auto;
    top: var(--zero);
    border-radius: 0 0 var(--radius-sm) var(--radius-sm);
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
    font-size: var(--fs-xs);
    margin: var(--space-1) 0 0;
  }
  .warn {
    color: var(--color-warning);
  }
</style>
