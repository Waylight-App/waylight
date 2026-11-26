<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import StatusDot from '../../../shared/ui/StatusDot.svelte';
  import { type ChoiceLoader, connectionIdOf, savedConnections } from '../../../shared/connection/connection';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import { grafanaSetup } from './grafanaSetup';
  import type { FoundService } from '../../../shared/setup/setup';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import type { GrafanaAlerts } from '../domain/grafana';
  import type { GrafanaUseCases } from '../usecase/grafanaUseCases';

  /** Props of the Grafana alerts widget (#574). */
  interface Props {
    /** Grafana found in the home network for the setup (#627). */
    discover?: () => Promise<FoundService[]>;
    grafana: GrafanaUseCases;
    config: Record<string, unknown>;
    /** Saved connections to offer in the setup, and storing the choice (#709). */
    connections?: ChoiceLoader;
    onsave?: (config: Record<string, unknown>) => Promise<void>;
  }
  let { grafana, discover, config, connections, onsave }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 60_000;
  let data = $state<GrafanaAlerts | null>(null);
  let error = $state('');
  let unconfigured = $state(false);

  const load = () =>
    grafana
      .alerts(connectionIdOf(config))
      .then((a) => ((data = a), (error = ''), (unconfigured = false)))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });
  $effect(() => poll(load, POLL_MS));
  // svelte-ignore state_referenced_locally
  const setup = grafanaSetup(grafana, discover);

  const since = (d: Date) =>
    d.toLocaleString(intlLocale(), { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
</script>

{#if unconfigured}
  <SetupInPlace
    text={t('grafana.connect_under_settings')}
    descriptor={setup}
    instance
    connections={savedConnections(['grafana'], connections, config, onsave)}
    ondone={() => void load().catch(() => undefined)}
  />
{:else if error && !data}
  <p class="error">{error}</p>
{:else if data}
  {#if data.alerts.length === 0}
    <p class="muted">{t('grafana.all_quiet')}</p>
  {:else}
    <ul>
      {#each data.alerts as a, i (a.name + i)}
        <ListRow>
          {#snippet lead()}<StatusDot level={a.state === 'pending' ? 'warning' : 'failed'} />{/snippet}
          <strong>{a.name}</strong>
          {#if a.state === 'pending'}<span class="muted"> · {t('grafana.pending')}</span>{/if}
          {#snippet sub()}
            {#if a.summary}<div>{a.summary}</div>{/if}
            {#if a.since}<div>{since(a.since)}</div>{/if}
          {/snippet}
        </ListRow>
      {/each}
    </ul>
  {/if}
  {#if data.url}<a class="small" href={data.url} target="_blank" rel="noopener noreferrer"
      >{t('grafana.alert_list')}</a
    >{/if}
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  ul {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
