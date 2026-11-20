<script lang="ts" generics="T extends { id: string }">
  import { t } from '../../../shared/i18n';
  import type { Snippet } from 'svelte';
  import { MAX_ROWS } from '../domain/layout';
  import {
    MOBILE_SIZES,
    type MobileItem,
    type MobileSize,
    moveMobile,
    sizeCells,
    updateMobile,
    usesGrid,
  } from '../domain/mobile';

  /**
   * Props of the phone layout (#92): one column, or a grid of four columns
   * once a widget has a size (#585), as in the mobile app.
   */
  interface Props {
    /** Phone layout in display order. */
    layout: MobileItem[];
    /** Widgets by ID; layout entries without a widget are skipped. */
    widgets: Map<string, T>;
    /** Edit mode: reorder, hide and resize. */
    editing: boolean;
    /** Whether the layout was edited (enables "reset"). */
    custom: boolean;
    /** Height of one grid row in pixels. */
    rowHeight?: number;
    /** Called with the changed layout. */
    onchange: (layout: MobileItem[]) => void;
    /** Derives the layout from the desktop grid again. */
    onreset: () => void;
    item: Snippet<[T]>;
    label: (item: T) => string;
  }
  let { layout, widgets, editing, custom, rowHeight = 56, onchange, onreset, item, label }: Props = $props();

  const visible = $derived(layout.filter((e) => widgets.has(e.id) && (editing || !e.hidden)));
  const grid = $derived(usesGrid(layout));
  let announcement = $state('');

  function move(e: MobileItem, delta: number, name: string) {
    const next = moveMobile(layout, e.id, delta);
    if (!next) return;
    announcement = t('dashboard.mobile_moved', { name, position: next.findIndex((i) => i.id === e.id) + 1 });
    onchange(next);
  }
</script>

{#if editing}
  <div class="mobile-bar">
    <p class="muted small">{t(custom ? 'dashboard.mobile_custom' : 'dashboard.mobile_derived')}</p>
    {#if custom}
      <button class="btn" type="button" onclick={onreset}>{t('dashboard.mobile_reset')}</button>
    {/if}
  </div>
{/if}

<div class="list" class:grid style:--row-height="{rowHeight}px">
  {#each visible as e, i (e.id)}
    {@const w = widgets.get(e.id)}
    {#if w}
      {@const name = label(w)}
      {@const cells = sizeCells(e.size)}
      <div
        class="cell"
        class:hidden={e.hidden}
        class:sized={cells !== undefined}
        style:--rows={e.rows}
        style:--span-cols={cells?.cols}
        style:--span-rows={cells?.rows}
        data-widget-id={e.id}
      >
        {#if editing}
          <div class="controls" role="group" aria-label={t('dashboard.mobile_controls', { name })}>
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('dashboard.mobile_up', { name })}
              disabled={i === 0}
              onclick={() => move(e, -1, name)}>↑</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('dashboard.mobile_down', { name })}
              disabled={i === visible.length - 1}
              onclick={() => move(e, 1, name)}>↓</button
            >
            <select
              class="input size"
              aria-label={t('dashboard.mobile_size', { name })}
              value={e.size ?? ''}
              onchange={(ev) =>
                onchange(
                  updateMobile(layout, e.id, {
                    size: (ev.currentTarget.value || undefined) as MobileSize | undefined,
                  }),
                )}
            >
              <option value="">{t('dashboard.mobile_size_full')}</option>
              {#each MOBILE_SIZES as s (s.size)}
                <option value={s.size}>{t(`dashboard.mobile_size_${s.size}`)}</option>
              {/each}
            </select>
            {#if !cells}
              <button
                class="btn btn-icon"
                type="button"
                aria-label={t('dashboard.mobile_shorter', { name })}
                disabled={e.rows <= 1}
                onclick={() => onchange(updateMobile(layout, e.id, { rows: e.rows - 1 }))}>−</button
              >
              <button
                class="btn btn-icon"
                type="button"
                aria-label={t('dashboard.mobile_taller', { name })}
                disabled={e.rows >= MAX_ROWS}
                onclick={() => onchange(updateMobile(layout, e.id, { rows: e.rows + 1 }))}>+</button
              >
            {/if}
            <label class="hide">
              <input
                type="checkbox"
                checked={e.hidden}
                onchange={(ev) => onchange(updateMobile(layout, e.id, { hidden: ev.currentTarget.checked }))}
              />
              {t('dashboard.mobile_hide')}
            </label>
          </div>
        {/if}
        <div class="content">{@render item(w)}</div>
      </div>
    {/if}
  {/each}
</div>
<p class="sr-only" aria-live="polite">{announcement}</p>

<style>
  .list {
    display: grid;
    gap: 12px;
  }
  /* The stored height is the minimum; long lists are capped and scroll. */
  .cell {
    display: flex;
    flex-direction: column;
    min-height: calc(var(--rows) * (var(--row-height) + 12px) - 12px);
    max-height: max(70vh, calc(var(--rows) * (var(--row-height) + 12px) - 12px));
    min-width: 0;
  }
  /* The grid of the mobile app (#585): four columns, square-ish cells;
     sized widgets span cells, the others take the full width. Dense packing
     fills gaps in display order. */
  .list.grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    /* A cell is at least one and a half rows high; full-width widgets
       without a size grow with their content as before. */
    grid-auto-rows: minmax(calc(var(--row-height) * 1.5), auto);
    grid-auto-flow: row dense;
  }
  .list.grid .cell {
    grid-column: 1 / -1;
  }
  .list.grid .cell.sized {
    grid-column: span var(--span-cols);
    grid-row: span var(--span-rows);
    min-height: 0;
    max-height: none;
  }
  .size {
    width: auto;
    padding-block: 0.15rem;
  }
  .content {
    flex: 1;
    display: flex;
    min-height: 0;
  }
  .content > :global(*) {
    flex: 1;
  }
  .cell.hidden .content {
    opacity: 0.4;
  }
  .controls,
  .mobile-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.25rem 0.5rem;
    margin-bottom: 0.25rem;
  }
  .mobile-bar {
    justify-content: space-between;
    margin-bottom: 0.75rem;
  }
  .hide {
    display: inline-flex;
    gap: 0.25rem;
    align-items: center;
  }
  .small {
    font-size: 0.85rem;
    margin: 0;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
</style>
