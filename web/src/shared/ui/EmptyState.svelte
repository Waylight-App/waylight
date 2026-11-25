<script lang="ts">
  import type { Snippet } from 'svelte';
  import { t } from '../i18n';
  import Icon from './icons/Icon.svelte';
  import type { IconName } from './icons/paths';
  import { useOpenSettings } from './settingsLink';
  import { useWidgetIcon } from './widgetIcon';

  /**
   * What a widget shows while it has nothing to show or is not set up
   * (#512, #737): a round symbol, one short sentence and one action. The
   * symbol is the widget's own unless one is given. The action is the button
   * that opens a settings section, or whatever the widget passes as children
   * (the setup of a connection). The layout follows the size of the tile, so
   * the sentence and the action stay whole even in the smallest one.
   */
  interface Props {
    text: string;
    /** Settings section that sets the widget up, e.g. "calendar". */
    section?: string;
    icon?: IconName;
    /** Actions that replace the settings button. */
    children?: Snippet;
  }
  let { text, section, icon, children }: Props = $props();
  const open = useOpenSettings();
  const widgetIcon = useWidgetIcon();
  const symbol = $derived(icon ?? widgetIcon ?? 'box');
</script>

<div class="empty">
  <span class="symbol"><Icon name={symbol} size={28} /></span>
  <p>{text}</p>
  {#if children}
    <div class="actions">{@render children()}</div>
  {:else if section && open}
    <div class="actions">
      <button class="btn btn-primary" type="button" onclick={() => open(section)}>{t('common.set_up')}</button
      >
    </div>
  {/if}
</div>

<style>
  /* The tile's content box is the size container (WidgetFrame), so the
     queries below follow the tile, not the screen. */
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    min-height: 100%;
    text-align: center;
    color: var(--color-text-subtle);
  }
  .symbol {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: none;
    width: 3.25rem;
    height: 3.25rem;
    border-radius: 50%;
    background: color-mix(in srgb, var(--color-accent) 14%, transparent);
    color: var(--color-accent-text);
  }
  p {
    margin: 0;
    max-width: 28ch;
    text-wrap: balance;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-2);
  }
  /* Low tiles lay the three parts side by side; without room for the
     symbol it goes, then the sentence is cut with an ellipsis before the
     action ever is. */
  @container widget (max-height: 150px) {
    .empty {
      flex-direction: row;
      justify-content: flex-start;
      text-align: left;
      gap: var(--space-3);
    }
    .symbol {
      width: 2.5rem;
      height: 2.5rem;
    }
    p {
      flex: 1;
      min-width: 0;
      display: -webkit-box;
      -webkit-line-clamp: 3;
      line-clamp: 3;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }
  }
  @container widget (max-height: 150px) and (max-width: 260px) {
    .empty {
      flex-wrap: wrap;
    }
    .symbol {
      display: none;
    }
  }
</style>
