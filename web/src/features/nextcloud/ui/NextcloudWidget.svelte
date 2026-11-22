<script lang="ts">
  import { type ChoiceLoader, connectionIdOf, savedConnections } from '../../../shared/connection/connection';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import { nextcloudSetup } from './nextcloudSetup';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatCount } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import { type NextcloudStatus, shortVersion } from '../domain/nextcloud';
  import type { NextcloudUseCases } from '../usecase/nextcloudUseCases';

  /** Props of the Nextcloud widget. */
  interface Props {
    nextcloud: NextcloudUseCases;
    config: Record<string, unknown>;
    /** Saved connections to offer in the setup, and storing the choice (#709). */
    connections?: ChoiceLoader;
    onsave?: (config: Record<string, unknown>) => Promise<void>;
  }
  let { nextcloud, config, connections, onsave }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 5 * 60_000;
  let status = $state<NextcloudStatus | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  const load = () =>
    nextcloud
      .status(connectionIdOf(config))
      .then((s) => ((status = s), (error = ''), (unconfigured = false)))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });
  $effect(() => poll(load, POLL_MS));
  // svelte-ignore state_referenced_locally
  const setup = nextcloudSetup(nextcloud);
</script>

{#if unconfigured}
  <SetupInPlace
    text={t('nextcloud.connect_nextcloud_under_settings_nextclo')}
    descriptor={setup}
    instance
    connections={savedConnections(['nextcloud'], connections, config, onsave)}
    ondone={() => void load().catch(() => undefined)}
  />
{:else if error && !status}
  <p class="error">{error}</p>
{:else if status}
  <p class="head">
    <strong>{t('nextcloud.nextcloud', { value: shortVersion(status.version) })}</strong>
    {#if status.maintenance}<span class="badge warn">{t('nextcloud.maintenance')}</span>{/if}
    {#if status.server?.appUpdates}<span class="badge"
        >{t('nextcloud.app_updates', { appUpdates: status.server.appUpdates })}</span
      >{/if}
  </p>
  {#if status.quota.total > 0}
    <Meter
      label={t('nextcloud.your_quota')}
      percent={status.quota.percent}
      detail="{formatBytes(status.quota.used)} of {formatBytes(status.quota.total)}"
    />
  {:else}
    <p class="muted small">{t('nextcloud.you_use_no_quota', { value: formatBytes(status.quota.used) })}</p>
  {/if}
  {#if status.server}
    <div class="numbers">
      <div>
        <strong>{formatBytes(status.server.freeSpace)}</strong><span class="muted">{t('nextcloud.free')}</span
        >
      </div>
      <div>
        <strong>{status.server.activeUsers1d}/{status.server.users}</strong><span class="muted"
          >{t('nextcloud.active_today')}</span
        >
      </div>
      <div>
        <strong>{formatCount(status.server.files)}</strong><span class="muted">{t('nextcloud.files')}</span>
      </div>
    </div>
  {:else if status.serverHint}
    <p class="muted small">{status.serverHint}</p>
  {/if}
  {#if status.error}<p class="error small">{status.error}</p>{/if}
{:else}
  <p class="muted">{t('nextcloud.loading')}</p>
{/if}

<style>
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin: 0 0 var(--space-2);
  }
  .badge {
    font-size: max(0.8rem, 0.75em); /* at least 12 px (#317) */
    padding: 1px 6px;
    border-radius: 999px;
    background: var(--color-surface-2);
  }
  .badge.warn {
    color: var(--color-warning);
  }
  .numbers {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-2);
    margin-top: var(--space-2);
    text-align: center;
  }
  .numbers div {
    display: flex;
    flex-direction: column;
  }
  .numbers span,
  .small {
    font-size: 0.8em;
  }
</style>
