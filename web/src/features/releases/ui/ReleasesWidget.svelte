<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { coverUrl, parseWidgetConfig, type ReleaseList } from '../domain/releases';
  import type { ReleasesUseCases } from '../usecase/releasesUseCases';

  /** Props of the new-releases widget (#63). */
  interface Props {
    releases: ReleasesUseCases;
    config: Record<string, unknown>;
  }
  let { releases, config }: Props = $props();
  const poll = usePolling();
  const weeks = $derived(parseWidgetConfig(config).weeks);

  let list = $state<ReleaseList | null>(null);
  let error = $state('');

  const date = (d: string) =>
    new Date(`${d}T00:00:00`).toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short' });

  $effect(() => {
    const w = weeks;
    return poll(
      () =>
        releases
          .list(w)
          .then((l) => ((list = l), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      30 * 60_000,
    );
  });
</script>

{#if error && !list}
  <p class="error">{error}</p>
{:else if list}
  <ul class="releases">
    {#each list.releases as r (r.id)}
      <ListRow nowrap>
        {#snippet lead()}<img src={coverUrl(r.id)} alt="" loading="lazy" />{/snippet}
        <strong title={r.title}>{r.title}</strong>
        {#snippet sub()}{r.artist} · {r.type} · {date(r.date)}{/snippet}
      </ListRow>
    {:else}
      <li class="muted">{t('releases.none', { weeks })}</li>
    {/each}
  </ul>
  {#if list.error}<p class="error small">{list.error}</p>{/if}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .releases {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  img {
    width: 3rem;
    height: 3rem;
    object-fit: cover;
    border-radius: var(--radius-sm);
    background: var(--color-border);
    flex: none;
  }
</style>
