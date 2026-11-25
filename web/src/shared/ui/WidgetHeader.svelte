<script lang="ts">
  import type { Snippet } from 'svelte';
  import Icon from './icons/Icon.svelte';
  import type { IconName } from './icons/paths';

  /**
   * Props of the one header every widget shows (#672): icon, title and an
   * optional action. The full card and the compact view both use it, so all
   * widgets name themselves the same way.
   */
  interface Props {
    title: string;
    icon?: IconName;
    /** Quiet text after the title, such as a count or a date. */
    meta?: string;
    /** Controls at the end of the line, such as the edit buttons. */
    action?: Snippet;
  }
  let { title, icon, meta, action }: Props = $props();
</script>

<div class="widget-header">
  <h3>
    {#if icon}<Icon name={icon} size={16} />{/if}
    <span class="title">{title}</span>
  </h3>
  {#if meta}<span class="meta">{meta}</span>{/if}
  {#if action}<span class="action">{@render action()}</span>{/if}
</div>

<style>
  /* Title, meta text and a compact view's status text share one baseline
     (#794): the title's own baseline is its text, not the icon's edge. */
  .widget-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-2);
    min-width: 0;
  }
  h3 {
    display: flex;
    align-items: baseline;
    gap: var(--space-1);
    min-width: 0;
    margin: 0;
    font-size: var(--fs-sm);
    font-weight: 600;
    color: var(--color-text);
  }
  /* 16px like every line icon; muted so the title leads. */
  h3 :global(svg) {
    flex: none;
    align-self: center;
    color: var(--color-text-muted);
  }
  .meta {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
    font-variant-numeric: tabular-nums;
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .action {
    display: flex;
    align-self: center;
    align-items: center;
    gap: var(--space-1);
    flex: none;
  }
</style>
