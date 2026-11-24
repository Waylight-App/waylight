<script lang="ts">
  import { type ChoiceLoader, connectionIdOf, savedConnections } from '../../../shared/connection/connection';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import { speedtestSetup } from './speedtestSetup';
  import { usePolling } from '../../../shared/refresh';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import LineChart from '../../../shared/ui/LineChart.svelte';
  import { type Range, type SpeedResults, formatMbps, ranges } from '../domain/speedtest';
  import type { SpeedtestUseCases } from '../usecase/speedtestUseCases';

  /** Props of the speedtest widget (#60). */
  interface Props {
    speedtest: SpeedtestUseCases;
    config: Record<string, unknown>;
    /** Saved connections to offer in the setup, and storing the choice (#709). */
    connections?: ChoiceLoader;
    onsave?: (config: Record<string, unknown>) => Promise<void>;
  }
  let { speedtest, config, connections, onsave }: Props = $props();
  const poll = usePolling();

  let range = $state<Range>(24);
  let data = $state<SpeedResults | null>(null);
  let error = $state('');
  let missing = $state(false);
  // Bumped after the setup assistant saved, so the effect loads again.
  let reload = $state(0);
  // svelte-ignore state_referenced_locally
  const setup = speedtestSetup(speedtest);

  $effect(() => {
    const hours = range;
    void reload;
    const load = () =>
      speedtest
        .results(hours, connectionIdOf(config))
        .then((d) => ((data = d), (error = ''), (missing = false)))
        .catch((err) => {
          missing = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
          throw err;
        });
    return poll(load, 10 * 60_000);
  });
</script>

{#if missing && !data}
  <SetupInPlace
    text={t('speedtest.not_set_up')}
    descriptor={setup}
    instance
    connections={savedConnections(['speedtest'], connections, config, onsave)}
    ondone={() => reload++}
  />
{:else if error && !data}
  <p class="error">{error}</p>
{:else if data}
  {#if data.latest}
    <div class="latest">
      <div>
        <span class="muted small">↓ {t('speedtest.download')}</span><strong
          >{formatMbps(data.latest.downloadMbps)}</strong
        ><span class="muted small">Mbit/s</span>
      </div>
      <div>
        <span class="muted small">↑ {t('speedtest.upload')}</span><strong
          >{formatMbps(data.latest.uploadMbps)}</strong
        ><span class="muted small">Mbit/s</span>
      </div>
      <div>
        <span class="muted small">{t('speedtest.ping')}</span><strong>{data.latest.pingMs.toFixed(0)}</strong
        ><span class="muted small">ms</span>
      </div>
    </div>
    <p class="muted small">{t('speedtest.measured', { time: data.latest.at.toLocaleString() })}</p>
  {:else}
    <p class="muted">{t('speedtest.no_results')}</p>
  {/if}
  <div class="ranges" role="group" aria-label={t('speedtest.range')}>
    {#each ranges as r (r)}
      <button class="btn" class:active={range === r} onclick={() => (range = r)} aria-pressed={range === r}
        >{t(r === 24 ? 'speedtest.range_day' : 'speedtest.range_week')}</button
      >
    {/each}
  </div>
  <LineChart
    times={data.history.map((m) => m.at)}
    series={[
      { label: t('speedtest.download'), values: data.history.map((m) => m.downloadMbps) },
      { label: t('speedtest.upload'), values: data.history.map((m) => m.uploadMbps) },
    ]}
    unit="Mbit/s"
    label={t('speedtest.history')}
  />
  {#if data.error}<p class="error small">{data.error}</p>{/if}
{:else}
  <p class="muted">{t('speedtest.loading')}</p>
{/if}

<style>
  .small {
    font-size: var(--fs-xs);
  }
  .latest {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-2);
  }
  .latest div {
    display: flex;
    flex-direction: column;
  }
  .latest strong {
    font-size: var(--fs-md);
    font-variant-numeric: tabular-nums;
  }
  .ranges {
    display: flex;
    gap: var(--space-1);
    margin: var(--space-1) 0;
  }
  .ranges .active {
    font-weight: 600;
  }
</style>
