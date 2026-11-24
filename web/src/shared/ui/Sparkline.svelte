<script lang="ts">
  import { sparklinePoints } from './sparkline';

  /** Props of a minimal line chart. */
  interface Props {
    values: number[];
    /** Fixed maximum (e.g. 100 for percentages); defaults to the largest value. */
    max?: number;
    label: string;
  }
  let { values, max, label }: Props = $props();
  const WIDTH = 100;
  const HEIGHT = 20;
  const points = $derived(sparklinePoints(values, WIDTH, HEIGHT, max));
</script>

{#if values.length > 1}
  <svg viewBox="0 0 {WIDTH} {HEIGHT}" preserveAspectRatio="none" role="img" aria-label={label}>
    <polyline
      {points}
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
      vector-effect="non-scaling-stroke"
    />
  </svg>
{/if}

<style>
  svg {
    display: block;
    width: 100%;
    height: clamp(14px, 4cqh, 32px);
    /* A neutral step: the accent colour is kept for the one series a chart is about. */
    color: var(--color-text-muted);
    opacity: 0.6;
  }
</style>
