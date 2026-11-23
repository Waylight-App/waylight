<script lang="ts">
  import { barHeights } from './bars';

  /** Props of a simple bar chart; a null value leaves its bar empty. */
  interface Props {
    values: (number | null)[];
    /** One text per bar for the tooltip and screen readers. */
    labels: string[];
    /** Short text under each bar, e.g. the month; optional. */
    ticks?: string[];
    label: string;
  }
  let { values, labels, ticks, label }: Props = $props();
  const b = $derived(barHeights(values.map((v) => v ?? 0)));
</script>

<ul class="chart" style:--zero="{b.zero}%" aria-label={label}>
  {#each values as v, i (i)}
    <li
      class="bar"
      class:empty={v === null}
      style:--h="{b.heights[i] ?? 0}%"
      title={labels[i]}
      aria-label={labels[i]}
    ></li>
  {/each}
</ul>
{#if ticks}
  <ul class="ticks" aria-hidden="true">
    {#each ticks as tick, i (i)}<li>{tick}</li>{/each}
  </ul>
{/if}

<style>
  /* One bar per value; the zero line sits at --zero from the top. */
  .chart,
  .ticks {
    display: flex;
    gap: var(--space-1);
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .chart {
    position: relative;
    height: 72px;
    border-bottom: 1px solid var(--color-border);
  }
  .bar {
    position: relative;
    flex: 1;
  }
  .bar::before {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: calc(100% - var(--zero));
    height: var(--h);
    background: color-mix(in srgb, var(--color-accent) 45%, transparent);
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  }
  .bar.empty::before {
    display: none;
  }
  .ticks li {
    flex: 1;
    text-align: center;
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
  }
</style>
