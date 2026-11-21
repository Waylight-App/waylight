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
    /** Controls at the end of the line, such as the edit buttons. */
    action?: Snippet;
  }
  let { title, icon, action }: Props = $props();
</script>

<div class="widget-header">
  <h3>
    {#if icon}<Icon name={icon} />{/if}
    <span class="title">{title}</span>
  </h3>
  {#if action}<span class="action">{@render action()}</span>{/if}
</div>

<style>
  .widget-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    min-width: 0;
  }
  h3 {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    min-width: 0;
    margin: 0;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  /* Sized in em so the icon follows the text size setting. */
  h3 :global(svg) {
    flex: none;
    width: 1.15em;
    height: 1.15em;
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .action {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    flex: none;
  }
</style>
