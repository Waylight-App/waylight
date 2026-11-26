<script lang="ts">
  import type { Snippet } from 'svelte';
  import { t } from '../../../shared/i18n';
  import { filterThemes, previewStyle, themeSides, type ThemeFilter, type ThemeInfo } from '../domain/themes';

  /**
   * Cards of designs, each previewed in its own colours; a click applies the
   * design. Shared by the setup wizard and the appearance settings (#754).
   */
  interface Props {
    themes: readonly ThemeInfo[];
    /** Id of the applied design. */
    selected: string;
    /** The filters on offer; one filter alone is not shown. */
    filters: readonly ThemeFilter[];
    /** Small cards in five columns, as in the wizard. */
    compact?: boolean;
    disabled?: boolean;
    onselect: (id: string) => void;
    /** A last tile in the grid, e.g. the import. */
    extra?: Snippet;
    /** Shown under the grid for the selected design. */
    detail?: Snippet<[ThemeInfo]>;
  }
  let {
    themes,
    selected,
    filters,
    compact = false,
    disabled = false,
    onselect,
    extra,
    detail,
  }: Props = $props();

  // svelte-ignore state_referenced_locally
  let filter = $state<ThemeFilter>(filters[0] ?? 'all');
  const shown = $derived(filterThemes(themes, filter));
  const current = $derived(themes.find((th) => th.id === selected));
  const filterLabel = (f: ThemeFilter) => t(`appearance.filter_${f}`);
</script>

{#if filters.length > 1}
  <div class="filters" role="group" aria-label={t('appearance.filter')}>
    {#each filters as f (f)}
      <button class="chip" type="button" aria-pressed={filter === f} onclick={() => (filter = f)}
        >{filterLabel(f)}</button
      >
    {/each}
  </div>
{/if}
<ul class="grid" class:compact>
  {#each shown as theme (theme.id)}
    {@const sides = themeSides(theme)}
    {@const on = selected === theme.id}
    <li>
      <button
        class="card"
        class:on
        type="button"
        style={previewStyle(theme, sides[0])}
        aria-pressed={on}
        aria-label={t('appearance.apply_theme', { name: theme.name })}
        {disabled}
        onclick={() => onselect(theme.id)}
      >
        <!-- One preview per side: light and dark next to each other for a design with both. -->
        <span class="previews" aria-hidden="true">
          {#each sides as side (side)}
            <span class="preview" style={previewStyle(theme, side)}>
              <span class="bar"></span>
              <span class="tile"
                ><span class="line"></span><span class="line short"></span><span class="dot"></span></span
              >
            </span>
          {/each}
        </span>
        <span class="name">{theme.name}</span>
        {#if on}<span class="check" aria-hidden="true">✓</span>{/if}
      </button>
    </li>
  {/each}
  {#if extra}<li>{@render extra()}</li>{/if}
</ul>
{#if current && detail}
  <div class="detail">{@render detail(current)}</div>
{/if}

<style>
  .filters {
    display: flex;
    gap: var(--space-2);
    margin: var(--space-2) 0;
  }
  .chip {
    padding: var(--space-1) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: var(--color-surface);
    color: var(--color-text);
    font: inherit;
    font-size: var(--fs-xs);
    cursor: pointer;
  }
  .chip[aria-pressed='true'] {
    background: var(--color-accent);
    border-color: var(--color-accent);
    color: var(--color-accent-contrast);
  }
  .grid {
    list-style: none;
    padding: 0;
    margin: var(--space-2) 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(9.5rem, 1fr));
    gap: var(--space-2);
  }
  .grid.compact {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
  /* The card renders with the design's own tokens: a live preview. */
  .card {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    width: 100%;
    height: 100%;
    padding: var(--space-2);
    border: 2px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    color: var(--color-text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  .card.on {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px var(--color-accent);
  }
  .card:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  .previews {
    display: flex;
    gap: var(--space-1);
  }
  .preview {
    flex: 1;
    min-width: 0;
    display: grid;
    gap: var(--space-1);
    padding: var(--space-1);
    border-radius: var(--radius-sm);
    background: var(--color-bg);
  }
  .bar {
    height: 6px;
    border-radius: var(--radius-sm);
    background: var(--color-surface-2, var(--color-surface));
  }
  .tile {
    display: grid;
    gap: var(--space-1);
    padding: var(--space-1);
    border-radius: var(--radius-sm);
    background: var(--color-surface);
    border: 1px solid var(--color-border);
  }
  .line {
    height: 5px;
    border-radius: var(--radius-sm);
    background: var(--color-text);
    opacity: 0.8;
  }
  .line.short {
    width: 60%;
    background: var(--color-text-subtle);
  }
  .dot {
    width: 30%;
    height: 8px;
    border-radius: var(--radius-sm);
    background: var(--color-accent);
  }
  .name {
    font-family: var(--font-display);
    font-size: var(--fs-xs);
    font-weight: 600;
  }
  .check {
    position: absolute;
    top: var(--space-1);
    right: var(--space-1);
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
    font-size: var(--fs-xs);
    font-weight: 600;
    line-height: 20px;
    text-align: center;
  }
  .detail {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-xs);
  }
</style>
