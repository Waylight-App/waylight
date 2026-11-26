<script lang="ts">
  import { useHeaderMeta } from '../../../shared/ui/widgetHeaderMeta';
  import { type ChoiceLoader, connectionIdOf, savedConnections } from '../../../shared/connection/connection';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import { speedtestSetup } from './speedtestSetup';
  import { usePolling } from '../../../shared/refresh';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
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

  // When the last measurement ran goes to the card header (#809).
  const setMeta = useHeaderMeta();
  $effect(() =>
    setMeta(
      data?.latest
        ? {
            text: t('speedtest.measured', {
              time: data.latest.at.toLocaleString(
                intlLocale(),
                // Today the clock time is enough; the header has little room.
                data.latest.at.toDateString() === new Date().toDateString()
                  ? { timeStyle: 'short' }
                  : { dateStyle: 'short', timeStyle: 'short' },
              ),
            }),
          }
        : null,
    ),
  );

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
        <span class="label">{t('speedtest.download')}</span>
        <span class="figure">{formatMbps(data.latest.downloadMbps)} <span class="unit">Mbit/s</span></span>
      </div>
      <div>
        <span class="label">{t('speedtest.upload')}</span>
        <span class="figure">{formatMbps(data.latest.uploadMbps)} <span class="unit">Mbit/s</span></span>
      </div>
      <div>
        <span class="label">{t('speedtest.ping')}</span>
        <span class="figure">{data.latest.pingMs.toFixed(0)} <span class="unit">ms</span></span>
      </div>
    </div>
  {:else}
    <p class="muted">{t('speedtest.no_results')}</p>
  {/if}
  <div class="bar">
    <div class="ranges" role="group" aria-label={t('speedtest.range')}>
      {#each ranges as r (r)}
        <button type="button" class="chip" aria-pressed={range === r} onclick={() => (range = r)}
          >{t(r === 24 ? 'speedtest.range_day' : 'speedtest.range_week')}</button
        >
      {/each}
    </div>
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
  /* Three stats above the chart: a quiet label, the figure in the display face with its unit inline. */
  .latest {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--space-4);
  }
  .latest div {
    display: grid;
  }
  .label,
  .unit {
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    color: var(--color-text-subtle);
  }
  .figure {
    font-family: var(--font-display);
    font-size: var(--fs-md);
    line-height: var(--lh-md);
    font-weight: 500;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }
  .unit {
    font-family: var(--font-sans);
    font-weight: 400;
  }
  .bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-2) var(--space-4);
    margin: var(--space-4) 0;
  }
  .ranges {
    display: flex;
    gap: var(--space-2);
  }
  /* 32 px chips of the guide; the chosen range is filled like the selected tab. */
  .chip {
    min-height: 32px;
    padding: 0 var(--space-3);
    border: 0;
    border-radius: 999px;
    background: none;
    color: inherit;
    font: inherit;
    font-size: var(--fs-xs);
    font-weight: 500;
  }
  .chip[aria-pressed='true'] {
    background: var(--color-surface-2);
    box-shadow: inset 0 0 0 1px var(--color-border-strong);
  }
</style>
