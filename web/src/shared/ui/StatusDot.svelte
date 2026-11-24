<script module lang="ts">
  /** The four states a status can have; the dot and its text always go together. */
  export type StatusLevel = 'ok' | 'warning' | 'failed' | 'stopped';
</script>

<script lang="ts">
  /**
   * Props of the status dot (design guide, Colour): a small dot plus a word,
   * so a status never rests on colour alone. `stopped` is a hollow ring, so
   * it differs from a filled dot by shape as well.
   */
  interface Props {
    level: StatusLevel;
    /** The word for the state. Without it the dot is decoration next to text. */
    label?: string;
    /** Draws the word; otherwise it stays for screen readers and the tooltip. */
    word?: boolean;
  }
  let { level, label, word = false }: Props = $props();
</script>

<span class="status" data-level={level} title={label && !word ? label : undefined}>
  <span class="dot" aria-hidden="true"></span>
  {#if label}<span class={word ? 'word' : 'sr-only'}>{label}</span>{/if}
</span>

<style>
  .status {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    flex: none;
  }
  .dot {
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    background: var(--color-ok);
  }
  [data-level='warning'] .dot {
    background: var(--color-warning);
  }
  [data-level='failed'] .dot {
    background: var(--color-danger);
  }
  [data-level='stopped'] .dot {
    background: transparent;
    box-shadow: inset 0 0 0 2px var(--color-text-muted);
  }
  .word {
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
  }
</style>
