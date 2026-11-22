<script lang="ts">
  import type { Snippet } from 'svelte';

  /**
   * A tile whose list scrolls while its main action stays in a fixed place at
   * the bottom (#737): a shopping list shows its add field however long the
   * list is. Without a footer the content simply fills the tile.
   */
  interface Props {
    children: Snippet;
    footer?: Snippet;
  }
  let { children, footer }: Props = $props();
</script>

<div class="tile">
  <div class="scroll">{@render children()}</div>
  {#if footer}<div class="footer">{@render footer()}</div>{/if}
</div>

<style>
  .tile {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    height: 100%;
    min-height: 0;
  }
  .scroll {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }
  .footer {
    flex: none;
  }
</style>
