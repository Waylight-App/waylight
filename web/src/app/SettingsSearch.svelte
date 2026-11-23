<script lang="ts">
  import { t } from '../shared/i18n';
  import { searchSettings, type SettingsHit, type SettingsIndex } from './settingsSearch';

  /** Props of the search field of the settings start page (#601). */
  interface Props {
    index: SettingsIndex;
    /** Called once when the field is first used, e.g. to load the other language. */
    onfirstuse?: () => void;
    onpick: (hit: SettingsHit) => void;
  }
  let { index, onfirstuse, onpick }: Props = $props();

  let query = $state('');
  let active = $state(0);
  let input = $state<HTMLInputElement>();
  let used = false;
  const hits = $derived(searchSettings(index, query));
  const open = $derived(query.trim() !== '');

  /** Focuses the field; the dialog calls it for the "/" shortcut. */
  export function focus() {
    input?.focus();
    input?.select();
  }

  function firstUse() {
    if (used) return;
    used = true;
    onfirstuse?.();
  }

  // The combobox keeps the focus in the field; arrows move the active option
  // (aria-activedescendant), Enter opens it, the first one by default.
  function key(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      if (hits.length === 0) return;
      active = (active + (e.key === 'ArrowDown' ? 1 : -1) + hits.length) % hits.length;
    } else if (e.key === 'Enter') {
      const hit = hits[active];
      if (!hit) return;
      e.preventDefault();
      onpick(hit);
    } else if (e.key === 'Escape') {
      // Clears the field first; a second Escape closes the dialog as usual.
      e.preventDefault();
      e.stopPropagation();
      query = '';
    }
  }

  const kindLabel = (hit: SettingsHit) =>
    hit.kind === 'area'
      ? t('settings.search.kind_area')
      : hit.kind === 'section'
        ? t('settings.search.kind_section')
        : t('settings.search.kind_field');
</script>

<div class="search">
  <input
    bind:this={input}
    bind:value={query}
    type="search"
    role="combobox"
    autocomplete="off"
    spellcheck="false"
    aria-label={t('settings.search.label')}
    placeholder={t('settings.search.placeholder')}
    aria-expanded={open}
    aria-controls="settings-search-results"
    aria-autocomplete="list"
    aria-activedescendant={open && hits[active] ? `settings-hit-${active}` : undefined}
    onfocus={firstUse}
    oninput={() => (active = 0)}
    onkeydown={key}
  />
  {#if open}
    <ul id="settings-search-results" role="listbox" aria-label={t('settings.search.results')}>
      {#each hits as hit, i (hit.key)}
        <!-- Keyboard use goes through the combobox field (arrows and Enter), as the ARIA pattern asks. -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <li
          id="settings-hit-{i}"
          role="option"
          aria-selected={i === active}
          class:active={i === active}
          onmousemove={() => (active = i)}
          onmousedown={(e) => e.preventDefault()}
          onclick={() => onpick(hit)}
        >
          <span class="title">{hit.title}</span>
          <span class="meta">{kindLabel(hit)}{hit.path ? ` · ${hit.path}` : ''}</span>
        </li>
      {/each}
    </ul>
  {/if}
  <!-- Announces the number of results while typing (#601). -->
  <p class="sr-only" aria-live="polite">
    {open
      ? hits.length === 0
        ? t('settings.search.none')
        : t(hits.length === 1 ? 'settings.search.count_one' : 'settings.search.count_other', {
            count: hits.length,
          })
      : ''}
  </p>
  {#if open && hits.length === 0}
    <p class="muted none">{t('settings.search.none')}</p>
  {/if}
</div>

<style>
  .search {
    margin-bottom: var(--space-4);
  }
  input {
    width: 100%;
    box-sizing: border-box;
  }
  ul {
    list-style: none;
    margin: var(--space-2) 0 0;
    padding: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
  }
  li {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-2) var(--space-3);
    cursor: pointer;
  }
  li + li {
    border-top: 1px solid var(--color-border);
  }
  li.active {
    background: color-mix(in srgb, var(--color-accent) 18%, transparent);
  }
  .title {
    font-weight: 600;
  }
  .meta {
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
  }
  .none {
    margin-top: var(--space-2);
  }
</style>
