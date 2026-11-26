<script lang="ts">
  import { useHeaderMeta } from '../../../shared/ui/widgetHeaderMeta';
  import { usePolling } from '../../../shared/refresh';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { formatBytes } from '../../../shared/ui/format';
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import StatusDot, { type StatusLevel } from '../../../shared/ui/StatusDot.svelte';
  import { type Disk, type Finding, isKnownFinding, overallStatus, powerOnAge } from '../domain/disk';
  import type { DiskHealthUseCases } from '../usecase/diskHealthUseCases';

  /** Props of the disk health widget (#59). */
  interface Props {
    diskhealth: DiskHealthUseCases;
  }
  let { diskhealth }: Props = $props();
  const poll = usePolling();

  let disks = $state<Disk[] | null>(null);
  let error = $state('');
  let missing = $state(false);

  $effect(() => {
    const load = () =>
      diskhealth
        .disks()
        .then((d) => ((disks = d), (error = ''), (missing = false)))
        .catch((err) => {
          missing = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
          throw err;
        });
    return poll(load, 5 * 60_000);
  });

  const describe = (f: Finding) =>
    isKnownFinding(f.code) ? t(`diskhealth.finding.${f.code}`, { value: f.value, name: f.name }) : f.message;
  const age = (h: number) => {
    const a = powerOnAge(h);
    return t('diskhealth.age', { years: a.years, days: a.days });
  };
  const levels: Record<string, StatusLevel> = { ok: 'ok', warning: 'warning', critical: 'failed' };
  // The second line of a row: the facts that matter at a glance, in one quiet line.
  const facts = (d: Disk) =>
    [
      d.device,
      d.temperature !== null ? `${d.temperature} °C` : '',
      d.percentageUsed !== null ? t('diskhealth.wear_short', { value: d.percentageUsed }) : '',
      d.reallocated ? t('diskhealth.reallocated_short', { n: d.reallocated }) : '',
      d.pending ? `${t('diskhealth.pending')} ${d.pending}` : '',
      d.powerOnHours !== null ? t('diskhealth.on_short', { age: age(d.powerOnHours) }) : '',
    ]
      .filter(Boolean)
      .join(' · ');
  const summary = $derived.by(() => {
    if (!disks) return null;
    const look = disks.filter((d) => d.status === 'warning' || d.status === 'critical');
    if (!look.length)
      return { text: t(`diskhealth.overall.${overallStatus(disks)}`), tone: 'neutral' as const };
    return {
      text: t(look.length === 1 ? 'diskhealth.needs_look' : 'diskhealth.need_look', { n: look.length }),
      tone: look.some((d) => d.status === 'critical') ? ('danger' as const) : ('warning' as const),
    };
  });
  // Shown in the card header (#809).
  const setMeta = useHeaderMeta();
  $effect(() => setMeta(summary));
</script>

{#if error && !disks}
  <p class={missing ? 'muted' : 'error'}>{error}</p>
{:else if disks}
  <ul class="disks">
    {#each disks as d (d.device)}
      <ListRow
        class="disk status-{d.status}"
        title={d.capacityBytes ? formatBytes(d.capacityBytes) : undefined}
      >
        {#snippet lead()}<span class="lead"><StatusDot level={levels[d.status] ?? 'stopped'} /></span
          >{/snippet}
        {d.model || d.device}
        {#snippet sub()}
          <span class="num">{facts(d)}</span>
          {#if d.findings.length}
            <ul class="findings">
              {#each d.findings as f, i (i)}<li class="status-{f.status}">{describe(f)}</li>{/each}
            </ul>
          {/if}
        {/snippet}
        {#snippet value()}<span class="state status-{d.status}">{t(`diskhealth.status.${d.status}`)}</span
          >{/snippet}
      </ListRow>
    {:else}
      <li class="muted">{t('diskhealth.no_disks')}</li>
    {/each}
  </ul>
{:else}
  <p class="muted">{t('diskhealth.loading')}</p>
{/if}

<style>
  .disks,
  .findings {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  /* The dot sits on the first line of a two-line row. */
  .disks :global(.disk) {
    align-items: flex-start;
  }
  .lead {
    display: flex;
    height: var(--lh-sm);
  }
  .num {
    font-variant-numeric: tabular-nums;
  }
  .state.status-warning,
  .findings li.status-warning {
    color: var(--color-warning);
  }
  .state.status-critical,
  .findings li.status-critical {
    color: var(--color-danger);
  }
  /* The findings explain a state in full: only a tall box has room for them. */
  .findings {
    display: none;
    margin-top: var(--space-1);
  }
  @container widget (min-height: 480px) {
    .findings {
      display: block;
    }
  }
</style>
