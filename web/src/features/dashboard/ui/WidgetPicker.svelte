<script lang="ts">
  import { t } from '../../../shared/i18n';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import type { Snippet } from 'svelte';
  import type { IconName } from '../../../shared/ui/icons/paths';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { groupByCategory, moveFocus, searchWidgetTypes } from '../domain/catalog';
  import type { WidgetType } from '../domain/dashboard';
  import { isCompactSize, WIDGET_SIZES, type WidgetSize } from '../domain/sizes';

  /** Props of the widget picker (#104). */
  interface Props {
    types: WidgetType[];
    icon: (kind: string) => IconName | undefined;
    /** Adds a widget: freely sized in its default size, or in a chosen size (#596). */
    onpick: (type: WidgetType, size?: WidgetSize) => void;
    /**
     * Example of a widget's compact view for the size previews, e.g.
     * "3 open · 1 overdue"; undefined for widgets without one, which show
     * their full view at every size. Example data, so previews work before
     * the widget is set up.
     */
    compactExample?: (kind: string) => string | undefined;
    onclose: () => void;
    /** Shown below the tiles, e.g. what a server adds in local mode (#291). */
    footer?: Snippet;
    /** Text the search starts with, e.g. "weather" from "add widget weather" (#476). */
    initialQuery?: string;
    /** Kinds that are already chosen or on the page; their tiles are marked. */
    selected?: ReadonlySet<string>;
    /** Offers the size previews next to each tile; the onboarding wizard has none. */
    sizes?: boolean;
  }
  let {
    types,
    icon,
    onpick,
    onclose,
    footer,
    initialQuery = '',
    selected = new Set<string>(),
    sizes = true,
    compactExample = () => undefined,
  }: Props = $props();

  // The widget whose sizes are shown instead of the list (#596).
  let sizing = $state<WidgetType | null>(null);
  let sizeStep = $state<HTMLElement>();
  $effect(() => {
    if (sizing) sizeStep?.querySelector<HTMLButtonElement>('button.preview')?.focus();
  });

  /** What a size shows of a widget: its compact example, or the full view. */
  function previewText(type: WidgetType, size: WidgetSize): string {
    const example = compactExample(type.kind);
    return example !== undefined && isCompactSize(size) ? example : t('picker.full_view');
  }

  // svelte-ignore state_referenced_locally
  let query = $state(initialQuery);
  let search = $state<HTMLInputElement>();
  const matches = $derived(searchWidgetTypes(types, query));
  // Ranked results keep their rank inside each category column.
  const groups = $derived(groupByCategory(matches));
  const tiles: HTMLButtonElement[][] = [];

  $effect(() => {
    if (!sizing) search?.focus();
  });

  function focusTile(c: number, r: number) {
    tiles[c]?.[r]?.focus();
  }

  function onTileKey(e: KeyboardEvent, c: number, r: number) {
    if (!['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(e.key)) return;
    e.preventDefault();
    if (e.key === 'ArrowUp' && c === 0 && r === 0) {
      search?.focus();
      return;
    }
    const [nc, nr] = moveFocus(
      groups.map((g) => g.types.length),
      [c, r],
      e.key as Parameters<typeof moveFocus>[2],
    );
    focusTile(nc, nr);
  }

  function onSearchKey(e: KeyboardEvent) {
    if (e.key === 'ArrowDown' && groups.length) {
      e.preventDefault();
      focusTile(0, 0);
    } else if (e.key === 'Enter' && matches[0]) {
      e.preventDefault();
      onpick(matches[0]);
    }
  }
</script>

<Modal title={t('dashboard.add_widget_2')} {onclose} size="wide">
  {#if sizing}
    {@const type = sizing}
    {@const ic = icon(type.kind)}
    <div class="sizes" bind:this={sizeStep}>
      <div class="sizes-head">
        <button class="btn" type="button" onclick={() => (sizing = null)}>← {t('picker.back')}</button>
        <h3 id="picker-size-title">{t('picker.choose_size', { name: type.title })}</h3>
      </div>
      <ul class="previews" aria-labelledby="picker-size-title">
        {#each WIDGET_SIZES as s (s.size)}
          {@const text = previewText(type, s.size)}
          {@const label = t(`dashboard.mobile_size_${s.size}`)}
          <li>
            <button
              class="preview"
              type="button"
              aria-label={t('picker.size_preview', { size: label, preview: text })}
              onclick={() => onpick(type, s.size)}
            >
              <span class="frame" style:--w={s.w} style:--h={s.h} aria-hidden="true">
                {#if isCompactSize(s.size) && compactExample(type.kind) !== undefined}
                  <span class="mini-title">{type.title}</span>
                  <span class="mini-state">{text}</span>
                {:else}
                  {#if ic}<span class="mini-icon"><Icon name={ic} size={18} /></span>{/if}
                  <span class="skeleton"></span>
                  <span class="skeleton short"></span>
                {/if}
              </span>
              <strong>{label}</strong>
              <span class="desc">{text}</span>
            </button>
          </li>
        {/each}
      </ul>
      <button class="btn" type="button" onclick={() => onpick(type)}>{t('picker.default_size')}</button>
    </div>
  {:else}
    <div class="picker">
      <input
        class="input search"
        type="search"
        bind:this={search}
        bind:value={query}
        onkeydown={onSearchKey}
        placeholder={t('picker.search')}
        aria-label={t('picker.search')}
        aria-controls="widget-picker-grid"
      />
      <p class="sr-only" role="status" aria-live="polite">{t('picker.count', { count: matches.length })}</p>
      {#if groups.length === 0}
        <p class="muted empty">{t('picker.no_match', { query })}</p>
      {:else}
        <div class="columns" id="widget-picker-grid" role="group" aria-label={t('picker.widgets')}>
          {#each groups as g, c (g.category)}
            <section class="column" aria-labelledby="picker-cat-{g.category}">
              <h3 id="picker-cat-{g.category}">{t(`picker.category.${g.category}`)}</h3>
              <ul>
                {#each g.types as type, r (type.kind)}
                  {@const ic = icon(type.kind)}
                  <li>
                    <button
                      class="tile"
                      class:selected={selected.has(type.kind)}
                      type="button"
                      bind:this={
                        () => tiles[c]?.[r],
                        (el) => {
                          tiles[c] ??= [];
                          if (el) tiles[c][r] = el;
                        }
                      }
                      onclick={() => onpick(type)}
                      onkeydown={(e) => onTileKey(e, c, r)}
                    >
                      {#if ic}<span class="icon"><Icon name={ic} size={28} /></span>{/if}
                      <strong>{type.title}</strong>
                      <span class="desc">{type.description}</span>
                      {#if selected.has(type.kind)}<span class="check" aria-hidden="true">✓</span><span
                          class="sr-only">{t('picker.selected')}</span
                        >{/if}
                    </button>
                    <!-- A click on the tile adds the widget at once; this opens the size previews. -->
                    {#if sizes}<button
                        class="sizes-button"
                        type="button"
                        aria-label={t('picker.choose_size', { name: type.title })}
                        title={t('picker.choose_size', { name: type.title })}
                        onclick={() => (sizing = type)}>⊞</button
                      >{/if}
                  </li>
                {/each}
              </ul>
            </section>
          {/each}
        </div>
      {/if}
      {#if footer}{@render footer()}{/if}
    </div>
  {/if}
</Modal>

<style>
  .picker {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    max-height: 70vh;
  }
  .search {
    width: 100%;
  }
  .columns {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--space-4);
    overflow-y: auto;
    padding: 2px;
  }
  h3 {
    margin: 0 0 var(--space-2);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-2);
  }
  .tile {
    display: grid;
    grid-template-columns: auto 1fr;
    grid-template-rows: auto auto;
    column-gap: var(--space-2);
    align-items: center;
    width: 100%;
    min-height: 4.5rem;
    padding: var(--space-2);
    padding-right: 36px;
    position: relative;
    text-align: left;
    border: 1px solid var(--color-border, rgba(127, 127, 127, 0.3));
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
    color: inherit;
    font: inherit;
    cursor: pointer;
    transition:
      border-color 0.1s,
      transform 0.1s;
  }
  .tile:hover {
    border-color: var(--color-accent);
  }
  .tile.selected {
    border-color: var(--color-accent);
    background: color-mix(in srgb, var(--color-accent) 14%, var(--color-surface-2));
  }
  /* The muted description falls below 4.5:1 on the tinted background. */
  .tile.selected .desc {
    color: var(--color-text);
  }
  .check {
    position: absolute;
    bottom: var(--space-1);
    right: var(--space-1);
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    font-size: 0.75rem;
    font-weight: 700;
    line-height: 20px;
    text-align: center;
  }
  .tile:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  .tile:active {
    transform: scale(0.98);
  }
  .icon {
    grid-row: 1 / span 2;
    display: flex;
    color: var(--color-accent-text);
  }
  .desc {
    grid-column: 2;
    font-size: 0.8rem;
    color: var(--color-text-muted);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
  li {
    position: relative;
  }
  .sizes-button {
    position: absolute;
    top: var(--space-1);
    right: var(--space-1);
    width: 28px;
    height: 28px;
    padding: 0;
    border: none;
    border-radius: var(--radius-small);
    background: transparent;
    color: var(--color-text-muted);
    cursor: pointer;
  }
  .sizes-button:hover,
  .sizes-button:focus-visible {
    color: var(--color-accent-text);
    background: color-mix(in srgb, var(--color-accent) 12%, transparent);
  }
  .sizes {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    align-items: flex-start;
  }
  .sizes-head {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .sizes-head h3 {
    margin: 0;
    font-size: 1rem;
    text-transform: none;
    letter-spacing: normal;
    color: var(--color-text);
  }
  .previews {
    width: 100%;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
  .preview {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    width: 100%;
    height: 100%;
    padding: var(--space-2);
    text-align: left;
    border: 1px solid var(--color-border, rgba(127, 127, 127, 0.3));
    border-radius: var(--radius-small);
    background: var(--color-surface-2);
    color: inherit;
    font: inherit;
    cursor: pointer;
  }
  .preview:hover {
    border-color: var(--color-accent);
  }
  .preview:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  /* The tile drawn in the proportions of its cells: a column is about one
     and a half times as wide as a row is high. */
  .frame {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 4px;
    width: calc(var(--w) / 6 * 100%);
    aspect-ratio: calc(var(--w) * 1.5) / var(--h);
    max-height: 120px;
    padding: 6px;
    overflow: hidden;
    border: 1px solid var(--color-border, rgba(127, 127, 127, 0.3));
    border-radius: var(--radius-small);
    background: var(--color-surface);
    font-size: 0.7rem;
  }
  .mini-title {
    color: var(--color-text-muted);
    text-transform: uppercase;
    font-size: 0.6rem;
  }
  .mini-title,
  .mini-state {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .mini-state {
    font-weight: 600;
  }
  .mini-icon {
    display: flex;
    color: var(--color-accent-text);
  }
  .skeleton {
    height: 4px;
    border-radius: 2px;
    background: var(--color-border, rgba(127, 127, 127, 0.3));
  }
  .skeleton.short {
    width: 60%;
  }
  .empty {
    text-align: center;
    padding: var(--space-6) 0;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }
  @media (max-width: 640px) {
    .picker {
      max-height: none;
    }
    .columns {
      grid-template-columns: 1fr;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .tile {
      transition: none;
    }
  }
</style>
