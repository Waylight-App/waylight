<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { Snippet } from 'svelte';
  import type { Widget } from '../domain/dashboard';
  import { GRID_COLUMNS } from '../domain/layout';
  import type { ZoneGroup } from '../domain/zones';

  /** Props of the zone layout of a tab (#619). */
  interface Props {
    groups: ZoneGroup[];
    /** Edit mode: empty zones stay visible as places to move widgets to. */
    editing: boolean;
    /** Renders one widget. */
    item: Snippet<[Widget]>;
    /** Height of one grid row in pixels; the same as the free grid's. */
    rowHeight?: number;
  }
  let { groups, editing, item, rowHeight = 56 }: Props = $props();
</script>

<!-- Zones keep each widget's cell size but not its place: widgets flow in
     the free grid's reading order, so nothing has to be arranged twice. -->
<div class="zones" style:--row-height="{rowHeight}px">
  {#each groups as group (group.zone)}
    <section class="zone" aria-labelledby="zone-{group.zone}" data-zone={group.zone}>
      <h2 id="zone-{group.zone}">{t(`dashboard.zone_${group.zone}`)}</h2>
      {#if group.widgets.length > 0}
        <div class="cells">
          {#each group.widgets as widget (widget.id)}
            <div
              class="cell"
              data-widget-id={widget.id}
              style:--w={Math.min(widget.w, GRID_COLUMNS)}
              style:--h={widget.h}
            >
              {@render item(widget)}
            </div>
          {/each}
        </div>
      {:else if editing}
        <p class="empty muted">{t('dashboard.zone_empty')}</p>
      {/if}
    </section>
  {/each}
</div>

<style>
  .zones {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }
  .zone h2 {
    margin: 0 0 var(--space-2);
    font-size: var(--fs-xs);
    font-weight: 600;
    color: var(--color-text-muted);
  }
  .cells {
    display: grid;
    grid-template-columns: repeat(12, minmax(0, 1fr));
    grid-auto-rows: var(--row-height);
    grid-auto-flow: row dense;
    gap: var(--space-3);
  }
  .cell {
    grid-column: span var(--w);
    grid-row: span var(--h);
    min-width: 0;
    min-height: 0;
  }
  .empty {
    margin: 0;
    padding: var(--space-3);
    border: 1px dashed var(--color-border);
    border-radius: var(--radius-md);
    text-align: center;
  }
  /* Zones stack on phones: one column, each widget the full width. */
  @media (max-width: 720px) {
    .cells {
      grid-template-columns: minmax(0, 1fr);
    }
    .cell {
      grid-column: 1;
    }
  }
</style>
