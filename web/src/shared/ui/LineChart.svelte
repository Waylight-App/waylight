<script module lang="ts">
  let charts = 0;
  const nextId = () => ++charts;
</script>

<script lang="ts">
  import { intlLocale } from '../i18n';
  import { areaPath, niceMax, type Point, smoothPath } from './chart';

  /** One line of the chart; the first series is the one the chart is about. */
  interface Series {
    label: string;
    values: number[];
  }

  /**
   * Props of the main chart of a screen (design guide, Charts): a smooth
   * line with a light fill, a muted grid and axis labels at --fs-xs. The
   * first series takes the accent, the others a neutral step.
   */
  interface Props {
    series: Series[];
    /** One time per value, oldest first; positions the points and labels the x axis. */
    times: Date[];
    /** Accessible description of the whole chart. */
    label: string;
    /** Unit written after the top value of the y axis. */
    unit?: string;
  }
  let { series, times, label, unit = '' }: Props = $props();

  const W = 100;
  const H = 40;
  const top = $derived(niceMax(Math.max(...series.flatMap((s) => s.values), 0)));
  const t0 = $derived(times[0]?.getTime() ?? 0);
  const span = $derived((times[times.length - 1]?.getTime() ?? 0) - t0 || 1);
  const lines = $derived(
    series.map((s) => {
      const pts: Point[] = s.values.map((v, i) => [
        (((times[i]?.getTime() ?? t0) - t0) / span) * W,
        H - (Math.min(Math.max(v, 0), top) / top) * H,
      ]);
      return { label: s.label, line: smoothPath(pts), area: areaPath(pts, H) };
    }),
  );

  // The area fill's gradient needs an id of its own per chart on the page.
  const fillId = `chart-fill-${nextId()}`;

  // Within a day the clock time is enough; longer ranges need the date.
  const sameDay = $derived(
    times.length > 1 && times[0]?.toDateString() === times[times.length - 1]?.toDateString(),
  );
  const stamp = (d: Date | undefined) =>
    d
      ? d.toLocaleString(
          intlLocale(),
          sameDay ? { timeStyle: 'short' } : { day: 'numeric', month: 'short', hour: '2-digit' },
        )
      : '';
</script>

{#if times.length > 1}
  <figure class="chart">
    <div class="plot">
      <div class="yaxis" aria-hidden="true">
        <span>{top}</span><span>0</span>
      </div>
      <svg viewBox="0 0 {W} {H}" preserveAspectRatio="none" role="img" aria-label={label}>
        <defs>
          <linearGradient id={fillId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stop-color="var(--color-accent-text)" stop-opacity="0.22" />
            <stop offset="1" stop-color="var(--color-accent-text)" stop-opacity="0" />
          </linearGradient>
        </defs>
        {#each [0, H / 3, (2 * H) / 3, H] as y (y)}
          <line class="grid" x1="0" x2={W} y1={y} y2={y} vector-effect="non-scaling-stroke" />
        {/each}
        {#each lines as l, i (l.label)}
          {#if i === 0}<path class="area" d={l.area} fill="url(#{fillId})" />{/if}
          <path class="line s{i === 0 ? 0 : 1}" d={l.line} vector-effect="non-scaling-stroke" />
        {/each}
      </svg>
      <div class="xaxis" aria-hidden="true">
        <span>{stamp(times[0])}</span><span>{stamp(times[times.length - 1])}</span>
      </div>
    </div>
    <figcaption>
      {#each series as s, i (s.label)}
        <span class="key s{i === 0 ? 0 : 1}"><span class="swatch" aria-hidden="true"></span>{s.label}</span>
      {/each}
      {#if unit}<span>{unit}</span>{/if}
    </figcaption>
  </figure>
{/if}

<style>
  .chart {
    margin: 0;
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
    font-variant-numeric: tabular-nums;
  }
  .plot {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--space-2);
  }
  .yaxis {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    font-family: var(--font-mono);
    line-height: 1;
    text-align: right;
  }
  svg {
    display: block;
    width: 100%;
    /* Grows with the widget's own box, like the rest of its content. */
    height: clamp(96px, 68cqh, 280px);
    overflow: visible;
  }
  .grid {
    stroke: var(--color-border);
    stroke-width: 1;
  }
  .line {
    fill: none;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .line.s0 {
    stroke: var(--color-accent-text);
  }
  .line.s1 {
    stroke: var(--color-text-subtle);
    stroke-width: 1.5;
    opacity: 0.7;
  }
  .area {
    stroke: none;
  }
  @media (max-width: 639px) {
    svg {
      height: 96px;
    }
  }
  .xaxis {
    display: flex;
    justify-content: space-between;
    /* Second column of the plot grid: lines up past the y labels. */
    grid-column: 2;
    font-family: var(--font-mono);
  }
  figcaption {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1) var(--space-4);
    margin-top: var(--space-1);
  }
  .key {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
  }
  .swatch {
    width: var(--space-4);
    border-top: 2px solid var(--color-accent-text);
  }
  .key.s1 .swatch {
    border-top-color: var(--color-text-subtle);
  }
</style>
