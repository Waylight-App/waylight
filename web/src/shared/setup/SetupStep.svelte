<script lang="ts">
  import type { Snippet } from 'svelte';

  /**
   * One numbered thing the setup asks for (#738): a label, a plain hint on
   * where to find it and the control. The label belongs to the control by
   * the id, so screen readers announce both.
   */
  interface Props {
    n: number;
    label: string;
    /** The id of the control the label names. */
    for?: string;
    hint?: Snippet;
    children: Snippet;
  }
  let { n, label, for: target, hint, children }: Props = $props();
</script>

<div class="step">
  <span class="num" aria-hidden="true">{n}</span>
  <div class="main">
    <label for={target}>{label}</label>
    {#if hint}<div class="hint">{@render hint()}</div>{/if}
    {@render children()}
  </div>
</div>

<style>
  .step {
    display: flex;
    gap: var(--space-4);
  }
  .num {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: none;
    width: 1.9rem;
    height: 1.9rem;
    margin-top: 2px;
    border-radius: 50%;
    background: color-mix(in srgb, var(--color-accent) 14%, transparent);
    color: var(--color-accent-text);
    font-weight: 700;
  }
  .main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  label {
    font-weight: 600;
  }
  .hint {
    color: var(--color-text-muted);
    font-size: 0.92em;
    line-height: 1.45;
    margin-bottom: var(--space-1);
  }
  .hint :global(ol) {
    margin: 0;
    padding-left: 1.3em;
  }
</style>
