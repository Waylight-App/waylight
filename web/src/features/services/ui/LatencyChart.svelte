<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { LatencyPoint } from '../domain/service';
  import { latencyShapes, latencySummary } from '../domain/latency';

  /** Props of the response time chart of a service tile (#862). */
  interface Props {
    name: string;
    points: LatencyPoint[];
  }
  let { name, points }: Props = $props();
  const WIDTH = 100;
  const HEIGHT = 20;
  const shapes = $derived(latencyShapes(points, WIDTH, HEIGHT));
  const summary = $derived(latencySummary(points));
  // Screen readers get the key values instead of the drawing.
  const label = $derived(
    summary.average === null
      ? t('services.latency_summary_none', { name, outages: summary.outages })
      : t('services.latency_summary', {
          name,
          latest: summary.latest ?? 0,
          average: summary.average,
          max: summary.max ?? 0,
          outages: summary.outages,
        }),
  );
</script>

<span class="chart" role="img" aria-label={label}>
  <svg viewBox="0 0 {WIDTH} {HEIGHT}" preserveAspectRatio="none" aria-hidden="true">
    {#each shapes.outages as o, i (i)}
      <rect class="outage" x={o.x} y="0" width={o.width} height={HEIGHT} />
    {/each}
    {#each shapes.lines as line, i (i)}
      <polyline
        points={line}
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        vector-effect="non-scaling-stroke"
      />
    {/each}
  </svg>
  <span class="caption" aria-hidden="true">
    <span>{t('services.latency_caption')}</span>
    <span class:down={summary.down}>
      {summary.down
        ? t('services.latency_down')
        : summary.latest === null
          ? ''
          : t('services.latency_ms', { value: summary.latest })}
    </span>
  </span>
</span>

<style>
  .chart {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }
  svg {
    display: block;
    width: 100%;
    height: clamp(14px, 4cqh, 32px);
    /* A neutral step: the accent colour is kept for the one series a chart is about. */
    color: var(--color-text-subtle);
  }
  polyline {
    opacity: 0.6;
  }
  /* Outages are bars in the danger colour, never part of the line. */
  .outage {
    fill: var(--color-danger);
    opacity: 0.5;
  }
  .caption {
    display: flex;
    justify-content: space-between;
    gap: var(--space-2);
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
    white-space: nowrap;
  }
  .caption > span:first-child {
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .down {
    color: var(--color-danger);
  }
</style>
