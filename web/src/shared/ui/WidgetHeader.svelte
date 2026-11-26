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
    /** Colour of the meta text: subtle by default, warning or danger when it needs a look. */
    metaTone?: 'neutral' | 'warning' | 'danger';
    /** A quiet control right after the title, such as the help button; it never pushes the meta text inwards. */
    titleAction?: Snippet;
    /** Controls at the end of the line, such as the edit buttons. */
    action?: Snippet;
  }
  let { title, icon, meta, metaTone = 'neutral', titleAction, action }: Props = $props();
</script>

<div class="widget-header">
  <h3>
    {#if icon}<Icon name={icon} size={16} />{/if}
    <span class="title">{title}</span>
  </h3>
  {#if titleAction}<span class="title-action">{@render titleAction()}</span>{/if}
  {#if meta}<span class="meta" data-tone={metaTone} title={meta}>{meta}</span>{/if}
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
    min-height: 24px;
  }
  h3 {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    min-width: 0;
    margin: 0;
    font-size: var(--fs-sm);
    line-height: var(--lh-sm);
    font-weight: 600;
    color: var(--color-text);
  }
  /* 16px like every line icon; subtle so the title leads. */
  h3 :global(svg) {
    flex: none;
    align-self: center;
    color: var(--color-text-subtle);
  }
  .meta {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    text-align: right;
    color: var(--color-text-subtle);
    font-variant-numeric: tabular-nums;
  }
  .meta[data-tone='warning'] {
    color: var(--color-warning);
  }
  .meta[data-tone='danger'] {
    color: var(--color-danger);
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .title-action {
    display: flex;
    align-self: center;
    flex: none;
    /* Without meta text it stays next to the title instead of floating to the middle. */
    margin: -4px auto -4px 0;
  }
  .action {
    display: flex;
    align-self: center;
    align-items: center;
    gap: var(--space-1);
    flex: none;
    /* Icon buttons are 32 px; they overhang the 24 px row and the card padding like in the guide. */
    margin: -4px -8px -4px 0;
  }
</style>
