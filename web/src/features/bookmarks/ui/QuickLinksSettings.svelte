<script lang="ts">
  import { t } from '../../../shared/i18n';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import type { Bookmark } from '../domain/bookmark';
  import {
    MAX_QUICK_LINKS,
    moveQuickLink,
    parseQuickLinksConfig,
    SHORTCUT_LINKS,
    toggleQuickLink,
  } from '../domain/quickLinks';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';

  /** Props of the quick links settings (#859): which bookmarks, in which order. */
  interface Props {
    bookmarks: BookmarksUseCases;
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { bookmarks, config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let ids = $state(parseQuickLinksConfig(config).ids);
  let all = $state<Bookmark[]>([]);

  $effect(() => {
    bookmarks.list().then((l) => (all = l.bookmarks));
  });

  const byId = $derived(new Map(all.map((b) => [b.id, b])));
  const chosen = $derived(ids.flatMap((id) => byId.get(id) ?? []));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ids: chosen.map((b) => b.id) });
  }}
>
  <p class="muted">{t('quicklinks.settings_hint', { max: MAX_QUICK_LINKS, keys: SHORTCUT_LINKS })}</p>

  {#if chosen.length > 0}
    <ol class="order" aria-label={t('quicklinks.order')}>
      {#each chosen as b, i (b.id)}
        <li>
          <span class="pos">{i + 1}</span>
          <span class="title">{b.title}</span>
          <button
            type="button"
            class="icon-btn"
            aria-label={t('quicklinks.move_up', { title: b.title })}
            disabled={i === 0}
            onclick={() => (ids = moveQuickLink(ids, b.id, -1))}><Icon name="chevron-up" size={16} /></button
          >
          <button
            type="button"
            class="icon-btn"
            aria-label={t('quicklinks.move_down', { title: b.title })}
            disabled={i === chosen.length - 1}
            onclick={() => (ids = moveQuickLink(ids, b.id, 1))}><Icon name="chevron" size={16} /></button
          >
        </li>
      {/each}
    </ol>
  {/if}

  <fieldset>
    <legend>{t('quicklinks.choose', { n: chosen.length, max: MAX_QUICK_LINKS })}</legend>
    {#if all.length === 0}
      <p class="muted">{t('quicklinks.empty')}</p>
    {/if}
    <ul class="choices">
      {#each all as b (b.id)}
        {@const on = ids.includes(b.id)}
        <li>
          <label>
            <input
              type="checkbox"
              checked={on}
              disabled={!on && ids.length >= MAX_QUICK_LINKS}
              onchange={() => (ids = toggleQuickLink(ids, b.id))}
            />
            {b.title}
          </label>
        </li>
      {/each}
    </ul>
  </fieldset>
  <button class="btn btn-primary" type="submit">{t('bookmarks.save')}</button>
</form>

<style>
  form {
    display: grid;
    gap: var(--space-3);
  }
  p {
    margin: 0;
  }
  ol,
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .order li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 32px;
  }
  .pos {
    width: 1.5rem;
    color: var(--color-text-subtle);
    font-size: var(--fs-xs);
  }
  .title {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .icon-btn {
    width: 32px;
    height: 32px;
    display: grid;
    place-items: center;
    border: 0;
    border-radius: var(--radius-sm);
    background: none;
    color: var(--color-text-subtle);
    cursor: pointer;
  }
  .icon-btn:hover:not(:disabled) {
    color: var(--color-text);
    background: var(--color-surface-2);
  }
  .icon-btn:disabled {
    opacity: 0.4;
    cursor: default;
  }
  fieldset {
    border: 0;
    margin: 0;
    padding: 0;
    min-width: 0;
  }
  legend {
    padding: 0;
    margin-bottom: var(--space-2);
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  .choices {
    display: grid;
    gap: var(--space-1);
    max-height: 18rem;
    overflow: auto;
  }
  .choices label {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
</style>
