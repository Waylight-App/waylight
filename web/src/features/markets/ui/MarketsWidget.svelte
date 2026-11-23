<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Sparkline from '../../../shared/ui/Sparkline.svelte';
  import {
    displayName,
    formatChange,
    formatValue,
    type MarketRow,
    parseConfig,
    relative,
  } from '../domain/markets';
  import type { MarketsUseCases } from '../usecase/marketsUseCases';

  /** Props of the markets widget (#284). */
  interface Props {
    markets: MarketsUseCases;
    config: Record<string, unknown>;
  }
  let { markets, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseConfig(config));
  let rows = $state<MarketRow[]>([]);
  let error = $state('');

  $effect(() => {
    const symbols = cfg.symbols;
    return poll(
      () =>
        markets
          .watchlist(symbols)
          .then((r) => ((rows = r), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      30 * 60_000,
    );
  });
</script>

{#if error && rows.length === 0}
  <p class="error">{error}</p>
{:else}
  <ul class="list">
    {#each rows as r (r.symbol)}
      {@const name = displayName(r.symbol)}
      {@const spark = relative(r.history)}
      <li>
        <span class="name">{name}</span>
        {#if r.value !== null}
          <span class="spark"
            ><Sparkline values={spark.values} max={spark.max} label={t('markets.trend', { name })} /></span
          >
          <span class="value">{formatValue(r.value, intlLocale())}</span>
          <span class="change" class:up={r.changePct > 0} class:down={r.changePct < 0}
            >{formatChange(r.changePct, intlLocale())}{#if r.stale}<span class="muted">
                · {t('markets.stale')}</span
              >{/if}</span
          >
        {:else}
          <span class="muted small err"
            >{t(r.error === 'no_data' ? 'markets.no_data' : 'markets.unavailable')}</span
          >
        {/if}
      </li>
    {/each}
  </ul>
  <p class="muted small">{t('markets.closes_note')}</p>
{/if}

<style>
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 64px auto auto;
    gap: var(--space-2);
    align-items: center;
    font-variant-numeric: tabular-nums;
  }
  .name {
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .value {
    text-align: right;
  }
  .change {
    text-align: right;
    font-size: var(--fs-xs);
    min-width: 5.5em;
  }
  .up {
    color: var(--color-success);
  }
  .down {
    color: var(--color-danger);
  }
  .err {
    grid-column: 2 / -1;
  }
  .small {
    font-size: var(--fs-xs);
    margin: var(--space-1) 0 0;
  }
</style>
