<script lang="ts">
  import { useHeaderMeta } from '../../../shared/ui/widgetHeaderMeta';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatBytes, formatUptime } from '../../../shared/ui/format';
  import Meter from '../../../shared/ui/Meter.svelte';
  import Sparkline from '../../../shared/ui/Sparkline.svelte';
  import { type HistoryPoint, type SystemSnapshot, daysUntilFull } from '../domain/system';
  import type { SystemUseCases } from '../usecase/systemUseCases';

  /** Props of the system widget. */
  interface Props {
    system: SystemUseCases;
  }
  let { system }: Props = $props();
  const poll = usePolling();

  /** The backend samples every 5 s. */
  const POLL_MS = 5000;
  let snap = $state<SystemSnapshot | null>(null);
  let history = $state<HistoryPoint[]>([]);
  /** History points are stored once a minute. */
  const HISTORY_POLL_MS = 60_000;
  let error = $state('');

  $effect(() => {
    const load = () =>
      system
        .snapshot()
        .then((s) => ((snap = s), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    const stopSnapshot = poll(load, POLL_MS);
    // The history is optional; failures still back off.
    const stopHistory = poll(() => system.history().then((h) => (history = h)), HISTORY_POLL_MS);
    return () => {
      stopSnapshot();
      stopHistory();
    };
  });

  // Host and uptime go to the card header (#809).
  const setMeta = useHeaderMeta();
  $effect(() =>
    setMeta(
      snap
        ? {
            text: [snap.hostname, t('system.up', { value: formatUptime(snap.uptimeSeconds) })]
              .filter(Boolean)
              .join(' · '),
          }
        : null,
    ),
  );

  const usage = (u: { used: number; total: number }) => `${formatBytes(u.used)} / ${formatBytes(u.total)}`;
</script>

{#if error && !snap}
  <p class="error">{error}</p>
{:else if snap}
  <div class="system">
    <div class="meters">
      <div>
        <Meter label={t('system.cpu_cores', { cpuCount: snap.cpuCount })} percent={snap.cpuPercent} />
        <span class="trend"
          ><Sparkline
            values={history.map((p) => p.cpu)}
            max={100}
            label={t('system.cpu_over_24_hours')}
          /></span
        >
      </div>
      <div>
        <Meter label={t('system.memory')} percent={snap.memory.percent} detail={usage(snap.memory)} />
        <span class="trend"
          ><Sparkline
            values={history.map((p) => p.memory)}
            max={100}
            label={t('system.memory_over_24_hours')}
          /></span
        >
      </div>
      {#if snap.swap.total > 0}<Meter
          label={t('system.swap')}
          percent={snap.swap.percent}
          detail={usage(snap.swap)}
        />{/if}
      {#each snap.disks as d (d.label)}
        {@const full = daysUntilFull(d.forecast)}
        <div>
          <Meter label={d.label} percent={d.usage.percent} detail={usage(d.usage)} />
          {#if full !== null}<p class="forecast muted" class:soon={full < 14}>
              {full === 0 ? t('system.forecast_full_now') : t('system.forecast_full_in', { days: full })}
            </p>{/if}
        </div>
      {/each}
      {#if snap.disks.length}
        <span class="trend"
          ><Sparkline
            values={history.map((p) => p.disk)}
            max={100}
            label={t('system.disk_usage_over_24_hours')}
          /></span
        >
      {/if}
    </div>
    <p class="facts muted">
      <span title={t('system.load_average_1_5_15_min')}
        >{t('system.load', { value: snap.load.map((l) => l.toFixed(2)).join(' · ') })}</span
      >
      {#if snap.temperature > 0}<span>{snap.temperature.toFixed(0)} °C</span>{/if}
    </p>
  </div>
{:else}
  <p class="muted">{t('system.loading')}</p>
{/if}

<style>
  .system {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  /* Label and value above a 4 px bar, 16 px apart (mockup composition). */
  .meters {
    display: grid;
    gap: var(--space-4);
  }
  .meters :global(.meter) {
    margin-bottom: 0;
  }
  .forecast {
    margin: var(--space-1) 0 0;
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
  }
  .forecast.soon {
    color: var(--color-warning);
  }
  .facts {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1) var(--space-3);
    margin: 0;
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    color: var(--color-text-subtle);
  }
  .facts {
    margin-top: auto;
  }
  /* The 24 hour trends are extra detail: only a tall box has room for them. */
  .trend {
    display: none;
  }
  @container widget (min-height: 520px) {
    .trend {
      display: block;
      margin-top: var(--space-2);
    }
  }
</style>
