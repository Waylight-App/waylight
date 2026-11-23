<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatCount } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import { type ImmichOverview, parseWidgetConfig, thumbnailUrl } from '../domain/immich';
  import type { ImmichUseCases } from '../usecase/immichUseCases';
  import { type ChoiceLoader, savedConnections } from '../../../shared/connection/connection';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import type { FoundService } from '../../../shared/setup/setup';
  import { immichSetup } from './immichSetup';

  /** Props of the Immich widget (#85). */
  interface Props {
    immich: ImmichUseCases;
    config: Record<string, unknown>;
    /** Services found in the home network for the setup (#590). */
    discover?: () => Promise<FoundService[]>;
    /** Saved connections to offer in the setup, and storing the choice (#709). */
    connections?: ChoiceLoader;
    onsave?: (config: Record<string, unknown>) => Promise<void>;
  }
  let { immich, config, discover, connections, onsave }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseWidgetConfig(config));

  const POLL_MS = 5 * 60_000;
  let overview = $state<ImmichOverview | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  const load = () =>
    immich
      .overview(cfg.connectionId)
      .then((o) => ((overview = o), (error = ''), (unconfigured = false)))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });
  $effect(() => poll(load, POLL_MS));
  // svelte-ignore state_referenced_locally
  const setup = immichSetup(immich, discover);
</script>

{#if unconfigured}
  <SetupInPlace
    text={t('immich.connect_under_settings')}
    descriptor={setup}
    instance
    connections={savedConnections(['immich'], connections, config, onsave)}
    ondone={() => void load().catch(() => undefined)}
  />
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  <div class="numbers">
    <div><strong>{formatCount(overview.photos)}</strong><span class="muted">{t('immich.photos')}</span></div>
    <div><strong>{formatCount(overview.videos)}</strong><span class="muted">{t('immich.videos')}</span></div>
  </div>
  {#if overview.diskTotal > 0}
    <Meter
      label={t('immich.storage')}
      percent={overview.diskPercent}
      detail="{formatBytes(overview.diskUsed)} / {formatBytes(overview.diskTotal)}"
    />
  {/if}
  {#if cfg.showMemories}
    {#each overview.memories as m (m.yearsAgo)}
      <h4>{t('immich.years_ago', { n: m.yearsAgo })}</h4>
      <ul class="thumbs">
        {#each m.assetIds as id, i (id + i)}
          <li>
            <img
              src={thumbnailUrl(id, cfg.connectionId)}
              alt={t('immich.memory_photo', { n: m.yearsAgo })}
              loading="lazy"
            />
          </li>
        {/each}
      </ul>
    {/each}
  {/if}
  {#if overview.version}<p class="muted small">Immich {overview.version}</p>{/if}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .numbers {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    text-align: center;
    margin-bottom: var(--space-2);
  }
  .numbers div {
    display: flex;
    flex-direction: column;
  }
  .numbers strong {
    font-size: clamp(1.1rem, 9cqmin, 2.5rem);
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
  }
  .thumbs {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(4rem, 1fr));
    gap: var(--space-1);
  }
  .thumbs img {
    width: 100%;
    aspect-ratio: 1;
    object-fit: cover;
    border-radius: var(--radius-sm);
    background: var(--color-border);
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
