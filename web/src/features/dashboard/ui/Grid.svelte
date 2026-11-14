<script lang="ts" generics="T extends LayoutItem">
  import { t } from '../../../shared/i18n';
  import { announce } from '../../../shared/a11y/live.svelte';
  import type { Snippet } from 'svelte';
  import {
    type ArrowKey,
    describeKeyboardStep,
    describeRect,
    GRID_COLUMNS,
    keyboardStep,
    type LayoutItem,
    moveItem,
    resizeItem,
    sameLayout,
  } from '../domain/layout';

  /** Props of the drag & drop grid. */
  interface Props {
    /** Items to render; their x/y/w/h are grid cells. */
    items: T[];
    /** Enables dragging and resizing. */
    editing: boolean;
    /** Height of one grid row in pixels. */
    rowHeight?: number;
    /** Gap between cells in pixels. */
    gap?: number;
    /** Called once after a drag or resize with the new layout. */
    onlayout: (items: LayoutItem[]) => void;
    /** Renders the content of one item; `handle` must be spread onto the drag handle. */
    item: Snippet<[T, { onpointerdown: (e: PointerEvent) => void }]>;
    /** Accessible name of an item (e.g. the widget title). */
    label?: (item: T) => string;
  }
  let {
    items,
    editing,
    rowHeight = 56,
    gap = 12,
    onlayout,
    item,
    label = () => t('dashboard.widget'),
  }: Props = $props();
  /** Arrow keys move, Shift + arrow keys resize the focused item. */
  function onkeydown(it: T, e: KeyboardEvent) {
    if (!editing || e.target !== e.currentTarget) return;
    if (!['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(e.key)) return;
    e.preventDefault();
    const base = items.map(({ id, x, y, w, h }) => ({ id, x, y, w, h }));
    const next = keyboardStep(base, it.id, e.key as ArrowKey, e.shiftKey);
    const r = (next ?? base).find((i) => i.id === it.id);
    if (!r) return;
    // Through the app's shared live region; mouse dragging stays silent.
    announce(describeKeyboardStep(label(it), r, e.shiftKey, next !== null), { repeat: true });
    if (next) onlayout(next);
  }

  let container: HTMLDivElement;
  let preview = $state<LayoutItem[] | null>(null);

  type Gesture = { mode: 'move' | 'resize'; id: string; startX: number; startY: number; origin: LayoutItem };
  let gesture = $state<Gesture | null>(null);

  const rects = $derived(new Map((preview ?? items).map((i) => [i.id, i])));

  function cellSize() {
    const width = container.clientWidth;
    return { col: (width - gap * (GRID_COLUMNS - 1)) / GRID_COLUMNS + gap, row: rowHeight + gap };
  }

  function start(mode: Gesture['mode'], target: T, e: PointerEvent) {
    if (!editing || e.button !== 0) return;
    e.preventDefault();
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    gesture = { mode, id: target.id, startX: e.clientX, startY: e.clientY, origin: { ...target } };
    preview = items.map(({ id, x, y, w, h }) => ({ id, x, y, w, h }));
  }

  function move(e: PointerEvent) {
    if (!gesture) return;
    const { col, row } = cellSize();
    const dx = Math.round((e.clientX - gesture.startX) / col);
    const dy = Math.round((e.clientY - gesture.startY) / row);
    const base = items.map(({ id, x, y, w, h }) => ({ id, x, y, w, h }));
    const o = gesture.origin;
    preview =
      gesture.mode === 'move'
        ? moveItem(base, o.id, o.x + dx, o.y + dy)
        : resizeItem(base, o.id, o.w + dx, o.h + dy);
  }

  function end() {
    if (!gesture) return;
    gesture = null;
    // A click without movement changes nothing (#272).
    if (preview && !sameLayout(items, preview)) onlayout(preview);
    preview = null;
  }
</script>

<div
  class="grid"
  class:editing
  bind:this={container}
  style:--row-height="{rowHeight}px"
  style:--gap="{gap}px"
  onpointermove={move}
  onpointerup={end}
  onpointercancel={end}
  role="presentation"
>
  {#each items as it (it.id)}
    {@const r = rects.get(it.id) ?? it}
    <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
    <div
      class="cell"
      data-widget-id={it.id}
      class:active={preview && gesture?.id === it.id}
      tabindex={editing ? 0 : undefined}
      role={editing ? 'group' : undefined}
      aria-roledescription={editing ? t('dashboard.movable_widget') : undefined}
      aria-label={editing
        ? t('dashboard.widget_group_label', { name: label(it), position: describeRect(r) })
        : undefined}
      onkeydown={(e) => onkeydown(it, e)}
      style:grid-area="{r.y + 1} / {r.x + 1} / span {r.h} / span {r.w}"
      style:--span={r.h}
    >
      {@render item(it, { onpointerdown: (e) => start('move', it, e) })}
      {#if editing}
        <button
          class="resize"
          type="button"
          aria-label={t('dashboard.resize_widget')}
          title={t('dashboard.resize_widget')}
          onpointerdown={(e) => start('resize', it, e)}
        ></button>
      {/if}
    </div>
  {/each}
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(12, minmax(0, 1fr));
    grid-auto-rows: var(--row-height);
    gap: var(--gap);
    min-height: 200px;
  }
  .grid.editing {
    background-image: linear-gradient(to right, transparent calc(100% - 1px), var(--color-border) 0);
    background-size: calc((100% + var(--gap)) / 12) 100%;
    touch-action: none;
    user-select: none;
  }
  .cell {
    position: relative;
    min-width: 0;
    min-height: 0;
  }
  .cell:focus-visible {
    outline: 3px solid var(--color-accent);
    outline-offset: 2px;
    border-radius: var(--radius);
  }
  .cell.active {
    z-index: 2;
    opacity: 0.85;
  }
  /* Hit area of at least 24 × 24 px (WCAG 2.5.8), 40 × 40 px on touch
     screens; the visible triangle stays small in the corner (#275). */
  .resize {
    position: absolute;
    right: 0;
    bottom: 0;
    width: 24px;
    height: 24px;
    padding: 0;
    border: none;
    background: none;
    cursor: nwse-resize;
    touch-action: none;
  }
  .resize::after {
    content: '';
    position: absolute;
    right: 2px;
    bottom: 2px;
    width: 16px;
    height: 16px;
    background: linear-gradient(135deg, transparent 50%, var(--color-accent) 50%);
    border-radius: 0 0 var(--radius) 0;
  }
  @media (pointer: coarse) {
    .resize {
      width: 40px;
      height: 40px;
    }
  }
</style>
