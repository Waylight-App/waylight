<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { GameDealsOverview } from '../domain/gamedeals';
  import type { GamedealsUseCases } from '../usecase/gamedealsUseCases';

  /** Props of the game deals widget (#62). */
  interface Props {
    gamedeals: GamedealsUseCases;
  }
  let { gamedeals }: Props = $props();
  const poll = usePolling();

  let overview = $state<GameDealsOverview | null>(null);
  let unconfigured = $state(false);
  let error = $state('');

  const date = (d: Date) => d.toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short' });

  $effect(() =>
    poll(
      () =>
        Promise.all([gamedeals.settings(), gamedeals.overview()])
          .then(([s, o]) => ((unconfigured = s.steamId === '' && !s.epic), (overview = o), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      60 * 60_000,
    ),
  );
</script>

{#if unconfigured}
  <p class="muted">{t('gamedeals.connect_under_settings')}</p>
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  {#if overview.freeNow.length || overview.freeNext.length}
    <h4>{t('gamedeals.free')}</h4>
    <ul class="list">
      {#each overview.freeNow as g (g.title)}
        <ListRow nowrap>
          <a href={g.url} target="_blank" rel="noopener noreferrer">{g.title}</a>
          {#snippet value()}<span class="small">{t('gamedeals.until', { date: date(g.end) })}</span>{/snippet}
        </ListRow>
      {/each}
      {#each overview.freeNext as g (g.title)}
        <ListRow nowrap class="next">
          <a href={g.url} target="_blank" rel="noopener noreferrer">{g.title}</a>
          {#snippet value()}<span class="small">{t('gamedeals.from', { date: date(g.start) })}</span
            >{/snippet}
        </ListRow>
      {/each}
    </ul>
  {/if}
  {#if overview.deals.length}
    <h4>{t('gamedeals.wishlist')}</h4>
    <ul class="list">
      {#each overview.deals as d (d.url)}
        <ListRow nowrap>
          {#snippet lead()}<span class="badge">−{d.discountPct} %</span>{/snippet}
          <a href={d.url} target="_blank" rel="noopener noreferrer">{d.name}</a>
          {#snippet value()}
            <span class="small"
              >{d.finalPrice}{d.endsAt ? ` · ${t('gamedeals.until', { date: date(d.endsAt) })}` : ''}</span
            >
          {/snippet}
        </ListRow>
      {/each}
    </ul>
  {/if}
  {#if !overview.deals.length && !overview.freeNow.length && !overview.freeNext.length}
    <p class="muted">{t('gamedeals.nothing')}</p>
  {/if}
  {#each overview.errors as e (e)}<p class="error small">
      {t('gamedeals.unreachable', { source: e })}
    </p>{/each}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  h4 {
    margin: 0 0 var(--space-1);
  }
  .list {
    list-style: none;
    padding: 0;
    margin: 0 0 var(--space-2);
  }
  :global(.row.next) a {
    opacity: 0.75;
  }
  .badge {
    background: var(--color-success);
    color: var(--color-accent-contrast);
    border-radius: var(--radius-sm);
    padding: 0 var(--space-1);
    font-size: var(--fs-xs);
    font-weight: 600;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
