<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
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
      <ListRow nowrap>
        <span class="name">{name}</span>
        {#snippet value()}
          {#if r.value !== null}
            <span class="quote">
              <span class="spark"
                ><Sparkline
                  values={spark.values}
                  max={spark.max}
                  label={t('markets.trend', { name })}
                /></span
              >
              <span class="num">{formatValue(r.value, intlLocale())}</span>
              <span class="change" class:up={r.changePct > 0} class:down={r.changePct < 0}
                >{formatChange(r.changePct, intlLocale())}{#if r.stale}<span class="muted">
                    · {t('markets.stale')}</span
                  >{/if}</span
              >
            </span>
          {:else}
            <span class="small">{t(r.error === 'no_data' ? 'markets.no_data' : 'markets.unavailable')}</span>
          {/if}
        {/snippet}
      </ListRow>
    {/each}
  </ul>
  <p class="muted small">{t('markets.closes_note')}</p>
{/if}

<style>
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  /* Figures stay on one line at every size; the name gives way first. */
  :global(.value:has(> .quote)) {
    max-width: none;
  }
  .quote {
    white-space: nowrap;
    overflow-wrap: normal;
    display: grid;
    grid-template-columns: 64px auto auto;
    gap: var(--space-2);
    align-items: center;
  }
  /* In a narrow tile the trend line goes before the figures get squeezed. */
  @container widget (max-width: 320px) {
    .spark {
      display: none;
    }
    .quote {
      grid-template-columns: auto auto;
    }
  }
  .name {
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .num {
    text-align: right;
    color: var(--color-text);
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
  .small {
    font-size: var(--fs-xs);
    margin: var(--space-1) 0 0;
  }
</style>
