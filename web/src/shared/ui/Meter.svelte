<script lang="ts">
  /** Props of a horizontal usage bar. */
  interface Props {
    label: string;
    /** 0–100 */
    percent: number;
    /** Text shown on the right, e.g. "3.1 / 7.6 GiB". */
    detail?: string;
    /**
     * Draws the bar in the accent: the one value of a group that needs
     * attention. Without it a bar turns hot from 60 %, so a widget does not
     * paint every bar amber.
     */
    hot?: boolean;
    /** A second, lighter segment after the fill (0–100), e.g. work still in progress. */
    extra?: number;
    /** Overrides the meter's accessible value text, e.g. "35 of 180 ECTS". */
    valueText?: string;
  }
  let { label, percent, detail = '', hot, extra = 0, valueText }: Props = $props();
  const clamp = (n: number) => Math.min(100, Math.max(0, n));
  const level = $derived(
    percent >= 90 ? 'danger' : percent >= 75 ? 'warning' : (hot ?? percent >= 60) ? 'hot' : 'ok',
  );
</script>

<div class="meter">
  <div class="row">
    <span>{label}</span>
    <span class="muted">{detail || `${percent.toFixed(0)} %`}</span>
  </div>
  <div
    class="bar"
    role="meter"
    aria-label={label}
    aria-valuenow={percent}
    aria-valuemin="0"
    aria-valuemax="100"
    aria-valuetext={valueText}
  >
    <div class="fill {level}" style:width="{clamp(percent)}%"></div>
    {#if extra > 0}<div
        class="fill extra {level}"
        style:width="{Math.min(extra, 100 - clamp(percent))}%"
      ></div>{/if}
  </div>
</div>

<style>
  .meter {
    margin-bottom: var(--space-4);
  }
  .row {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .row .muted {
    line-height: var(--lh-xs);
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    font-variant-numeric: tabular-nums;
  }
  /* A thin pill; the fill is quiet until a level asks for attention. */
  .bar {
    height: 4px;
    border-radius: 999px;
    display: flex;
    background: var(--color-surface-2);
    overflow: hidden;
  }
  .fill {
    height: 100%;
    border-radius: 999px;
    background: var(--color-text-muted);
    transition: width 400ms ease;
  }
  .fill.hot {
    background: var(--color-accent);
  }
  .fill.warning {
    background: var(--color-warning);
  }
  .fill.danger {
    background: var(--color-danger);
  }
  .fill.extra {
    opacity: 0.35;
  }
</style>
