<script lang="ts">
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { t } from '../../../shared/i18n';
  import type { Snippet } from 'svelte';
  import WidgetHeader from '../../../shared/ui/WidgetHeader.svelte';
  import { provideWidgetIcon } from '../../../shared/ui/widgetIcon';
  import { provideWidgetKind } from '../../../shared/setup/walk.svelte';
  import type { IconName } from '../../../shared/ui/icons/paths';
  import { WIDGET_SIZES, type WidgetSize } from '../domain/sizes';
  import { ZONES, type Zone } from '../domain/zones';

  /** Props of the card that wraps every widget. */
  interface Props {
    title: string;
    /** The widget kind; the setup walk of the onboarding finds its widget by it. */
    kind?: string;
    /** Icon of the widget kind, shown before the title. */
    icon?: IconName;
    editing: boolean;
    /** Spread onto the title bar so it acts as drag handle in edit mode. */
    handle: { onpointerdown: (e: PointerEvent) => void };
    /** Opens the widget settings; omitted for widgets without settings. */
    onsettings?: () => void;
    /** Opens the short help of the widget kind (#678); omitted where none is offered. */
    onhelp?: () => void;
    ondelete: () => void;
    /** Copies the widget including its configuration. */
    onduplicate?: () => void;
    /** Hides the title bar outside edit mode (e.g. for the search bar). */
    chromeless?: boolean;
    /**
     * Shows the compact view (#595): no title bar outside edit mode, since
     * the compact view names the widget itself.
     */
    compact?: boolean;
    /** Opens the detail panel; makes the whole card a button outside edit mode. */
    onopen?: (opener: HTMLElement) => void;
    /** Size of the tile (#595); with onsize the edit bar offers a size menu. */
    size?: WidgetSize;
    onsize?: (size: WidgetSize | undefined) => void;
    /** Zone of the tile on a zoned tab (#619); with onzone the edit bar offers a zone menu. */
    zone?: Zone;
    onzone?: (zone: Zone) => void;
    /** Reports the size of the content box, so the view can fall back to compact. */
    onbox?: (width: number, height: number) => void;
    children: Snippet;
  }
  let {
    title,
    kind,
    icon,
    editing,
    handle,
    onsettings,
    onhelp,
    ondelete,
    onduplicate,
    chromeless = false,
    compact = false,
    onopen,
    size,
    onsize,
    zone,
    onzone,
    onbox,
    children,
  }: Props = $props();

  provideWidgetIcon(() => icon);
  provideWidgetKind(() => kind);
  let content = $state<HTMLElement>();
  $effect(() => {
    const el = content;
    const report = onbox;
    if (!el || !report || typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(([entry]) => {
      if (entry) report(entry.contentRect.width, entry.contentRect.height);
    });
    observer.observe(el);
    return () => observer.disconnect();
  });
</script>

{#snippet editActions()}
  {#if onsize}
    <select
      class="input size"
      aria-label={t('dashboard.widget_size', { name: title })}
      title={t('dashboard.widget_size', { name: title })}
      value={size ?? ''}
      onpointerdown={(e) => e.stopPropagation()}
      onchange={(e) => onsize((e.currentTarget.value || undefined) as WidgetSize | undefined)}
    >
      <option value="">{t('dashboard.size_free')}</option>
      {#each WIDGET_SIZES as s (s.size)}
        <option value={s.size}>{t(`dashboard.mobile_size_${s.size}`)}</option>
      {/each}
    </select>
  {/if}
  {#if onzone}
    <select
      class="input size"
      aria-label={t('dashboard.widget_zone', { name: title })}
      title={t('dashboard.widget_zone', { name: title })}
      value={zone}
      onpointerdown={(e) => e.stopPropagation()}
      onchange={(e) => onzone(e.currentTarget.value as Zone)}
    >
      {#each ZONES as z (z)}
        <option value={z}>{t(`dashboard.zone_${z}`)}</option>
      {/each}
    </select>
  {/if}
  {#if onsettings}
    <button
      class="btn btn-icon"
      type="button"
      aria-label={t('dashboard.widget_settings')}
      title={t('dashboard.widget_settings')}
      onpointerdown={(e) => e.stopPropagation()}
      onclick={onsettings}><Icon name="gear" size={16} /></button
    >
  {/if}
  {#if onduplicate}
    <button
      class="btn btn-icon"
      type="button"
      aria-label={t('dashboard.duplicate_widget')}
      title={t('dashboard.duplicate_widget')}
      onpointerdown={(e) => e.stopPropagation()}
      onclick={onduplicate}>⧉</button
    >
  {/if}
  <button
    class="btn btn-icon"
    type="button"
    aria-label={t('dashboard.remove_widget')}
    title={t('dashboard.remove_widget')}
    onpointerdown={(e) => e.stopPropagation()}
    onclick={ondelete}>✕</button
  >
{/snippet}

{#snippet headerActions()}
  {#if onhelp}
    <button
      class="btn btn-icon help"
      class:editing
      type="button"
      aria-haspopup="dialog"
      aria-label={t('dashboard.widget_help', { name: title })}
      title={t('dashboard.widget_help', { name: title })}
      onpointerdown={(e) => e.stopPropagation()}
      onclick={onhelp}>?</button
    >
  {/if}
  {#if editing}{@render editActions()}{/if}
{/snippet}

<article
  class="widget card"
  class:chromeless={chromeless && !editing}
  class:compact={compact && !editing}
  class:openable={!!onopen && !editing}
>
  {#if editing || (!chromeless && !compact)}
    <!-- Dragging is pointer-only; settings and remove buttons stay keyboard accessible. -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <header class:draggable={editing} onpointerdown={handle.onpointerdown}>
      <WidgetHeader {title} {icon} action={onhelp || editing ? headerActions : undefined} />
    </header>
  {/if}
  <div class="content" bind:this={content}><div class="body">{@render children()}</div></div>
  {#if onopen && !editing}
    <!-- The whole card opens the details; a stretched button keeps one tab stop
         and leaves room for quick actions above it (#598). -->
    <button
      class="open"
      type="button"
      aria-haspopup="dialog"
      aria-label={t('dashboard.open_details', { name: title })}
      onclick={(e) => onopen(e.currentTarget)}
    ></button>
  {/if}
</article>

<style>
  .widget {
    position: relative;
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    opacity: var(--widget-opacity);
    /* A 1 px hairline sets the card off (#792); .card draws the border. */
    border-color: var(--color-widget-border);
  }
  .widget.chromeless {
    background: transparent;
    border-color: transparent;
    box-shadow: none;
  }
  header {
    padding: var(--space-2) var(--space-3) 0;
    min-height: 32px;
    display: flex;
    flex-direction: column;
    justify-content: center;
  }
  header.draggable {
    cursor: grab;
  }
  /* The help stays quiet until the card is hovered or the button focused,
     so a dashboard of widgets does not show a row of question marks. */
  .help {
    color: var(--color-text-muted);
    opacity: 0;
    transition: opacity 0.15s;
  }
  .widget:hover .help,
  .help.editing,
  .help:focus-visible {
    opacity: 1;
  }
  @media (hover: none) {
    .help {
      opacity: 1;
    }
  }
  /* The content area is a size container: widgets size text and layout
     relative to their own box (cqi/cqh/cqmin, @container widget …) instead
     of the viewport, so they scale when the widget is resized. */
  .content {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: var(--space-2) var(--space-3) var(--space-3);
    container: widget / size;
  }
  .body {
    height: 100%;
    font-size: clamp(0.85rem, 0.7rem + 1.2cqmin, 1.3rem);
  }
  .compact .content {
    padding: var(--space-2) var(--space-3);
  }
  /* Clicks fall through the content to the stretched "show details" button
     below it; only controls that opt in with pointer-events: auto (quick
     actions, #598) take them. */
  .openable .content {
    position: relative;
    z-index: 2;
    pointer-events: none;
  }
  .open {
    position: absolute;
    inset: 0;
    z-index: 1;
    padding: 0;
    border: none;
    border-radius: inherit;
    background: transparent;
    cursor: pointer;
  }
  .open:hover {
    background: color-mix(in srgb, var(--color-accent) 6%, transparent);
  }
  .open:focus-visible {
    outline: 3px solid var(--color-accent-text);
    outline-offset: -3px;
  }
  .size {
    width: auto;
    padding: var(--space-1) var(--space-1);
    font-size: var(--fs-xs);
  }
  .chromeless .content {
    padding: 0;
    overflow: visible;
  }
  .chromeless .body {
    height: 100%;
  }
  /* Cards on a phone use 16 px padding (#771). */
  @media (max-width: 639px) {
    header {
      padding: var(--space-4) var(--space-4) 0;
    }
    .content {
      padding: var(--space-2) var(--space-4) var(--space-4);
    }
    .compact .content {
      padding: var(--space-3) var(--space-4);
    }
  }
  /* Stacked on phones the height comes from the content, so only the width
     can be used as container size. */
  @media (max-width: 720px) {
    .content {
      container-type: inline-size;
    }
    .body {
      height: auto;
    }
  }
</style>
