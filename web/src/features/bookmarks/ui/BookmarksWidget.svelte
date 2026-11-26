<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import {
    type Bookmark,
    iconSrc,
    initials,
    linkProblem,
    parseBookmarksConfig,
    sortBookmarks,
  } from '../domain/bookmark';
  import type { BookmarkList, BookmarksUseCases } from '../usecase/bookmarksUseCases';
  import BookmarkForm from './BookmarkForm.svelte';
  import { useHeaderAction } from '../../../shared/ui/widgetHeaderAction';

  /** Props of the bookmarks widget. */
  interface Props {
    bookmarks: BookmarksUseCases;
    config: Record<string, unknown>;
  }
  let { bookmarks, config }: Props = $props();

  const cfg = $derived(parseBookmarksConfig(config));
  let data = $state<BookmarkList | null>(null);
  let query = $state('');
  let error = $state('');
  let editing = $state<Bookmark | 'new' | null>(null);
  // "Broken links only" (#72); offered once a check found a broken link.
  let brokenOnly = $state(false);
  const anyBroken = $derived(data?.bookmarks.some((b) => b.link.state === 'broken') ?? false);

  // A short list needs no filter field; it appears once the list is long (#817).
  const FILTER_FROM = 8;
  let total = $state(0);
  const filterable = $derived(total >= FILTER_FROM || query !== '' || brokenOnly);

  async function load() {
    try {
      data = await bookmarks.list({ folderId: cfg.folderId, tag: cfg.tag, query, broken: brokenOnly });
      if (query === '' && !brokenOnly) total = data.bookmarks.length;
      error = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void cfg;
    void query;
    void brokenOnly;
    load();
  });

  // The add button sits in the card header (#809).
  const headerAction = useHeaderAction();
  $effect(() =>
    headerAction.set({ icon: 'plus', label: t('bookmarks.add_bookmark'), onclick: () => (editing = 'new') }),
  );

  async function remove(b: Bookmark) {
    if (!confirm(t('common.confirm_delete', { name: b.title }))) return;
    await bookmarks.delete(b.id).catch((err) => (error = errorMessage(err)));
    editing = null;
    load();
  }
</script>

{#if filterable || anyBroken || !headerAction.inHeader}
  <div class="toolbar">
    {#if filterable}<input
        class="input"
        type="search"
        placeholder={t('bookmarks.filter')}
        bind:value={query}
        aria-label={t('bookmarks.filter_bookmarks')}
      />{/if}
    {#if anyBroken || brokenOnly}
      <button
        class="btn btn-icon"
        type="button"
        aria-pressed={brokenOnly}
        title={t('bookmarks.broken_only')}
        aria-label={t('bookmarks.broken_only')}
        onclick={() => (brokenOnly = !brokenOnly)}><Icon name="warning" size={16} /></button
      >
    {/if}
    {#if !headerAction.inHeader}
      <button
        class="btn btn-icon"
        type="button"
        aria-label={t('bookmarks.add_bookmark')}
        onclick={() => (editing = 'new')}><Icon name="plus" size={16} /></button
      >
    {/if}
  </div>
{/if}
{#if error}<p class="error">{error}</p>{/if}
{#if data}
  <ul class="list">
    {#each sortBookmarks(data.bookmarks, cfg.sort, intlLocale()) as b (b.id)}
      {@const icon = iconSrc(b)}
      {@const problem = linkProblem(b.link, intlLocale())}
      <ListRow nowrap showValue={!!problem || b.tags.length > 0}>
        <!-- The tile shows the site's icon; the initials under it show when there is none. -->
        {#snippet lead()}<span class="tile" aria-hidden="true"
            >{initials(b.title)}{#if icon}<img
                src={icon}
                alt=""
                width="16"
                height="16"
                loading="lazy"
                referrerpolicy="no-referrer"
                onerror={(e) => (e.currentTarget as HTMLImageElement).remove()}
              />{/if}</span
          >{/snippet}
        <a href={b.url} target="_blank" rel="noopener noreferrer" title={b.description || b.url}>
          <span class="title">{b.title}</span>
        </a>
        {#snippet value()}
          {#if b.tags.length > 0}<span class="tags">{b.tags.join(' ')}</span>{/if}
          {#if problem}<span
              class="link-state"
              class:broken={b.link.state === 'broken'}
              title={problem}
              aria-label={problem}
              role="img"
              >{#if b.link.state === 'broken'}<Icon name="warning" size={16} />{:else}↪{/if}</span
            >{/if}
        {/snippet}
        {#snippet trail()}
          {#if b.mine}
            <span class="tools">
              <button
                class="btn btn-icon"
                type="button"
                aria-label={t('bookmarks.edit', { title: b.title })}
                onclick={() => (editing = b)}><Icon name="pencil" size={16} /></button
              >
            </span>
          {:else}
            <!-- Shared links (#298) are read-only and marked as such. -->
            <span
              class="shared"
              title={t('bookmarks.shared_with_you')}
              aria-label={t('bookmarks.shared_with_you')}
              role="img"><Icon name="users" size={16} /></span
            >
          {/if}
        {/snippet}
      </ListRow>
    {:else}
      <li class="muted">
        {brokenOnly ? t('bookmarks.no_broken') : query ? t('bookmarks.none_match') : t('bookmarks.none_yet')}
      </li>
    {/each}
  </ul>
{/if}

{#if editing && data}
  <Modal
    title={editing === 'new' ? t('bookmarks.add_bookmark') : t('bookmarks.edit_bookmark')}
    onclose={() => (editing = null)}
  >
    <BookmarkForm
      {bookmarks}
      folders={data.folders.filter((f) => f.mine)}
      bookmark={editing === 'new' ? undefined : editing}
      folderId={cfg.folderId === 'none' ? '' : cfg.folderId}
      onsaved={() => {
        editing = null;
        load();
      }}
    />
    {#if editing !== 'new'}
      {@const b = editing}
      <p>
        <button class="btn btn-danger" type="button" onclick={() => remove(b)}
          >{t('bookmarks.delete_bookmark')}</button
        >
      </p>
    {/if}
  </Modal>
{/if}

<style>
  .toolbar {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  a {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 0;
    color: inherit;
    text-decoration: none;
  }
  a:hover .title {
    text-decoration: underline;
  }
  .title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .link-state {
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  .link-state.broken {
    color: var(--color-danger);
  }
  /* 32 px tile like the guide's bookmark rows: icon, or initials when there is none. */
  .tile {
    position: relative;
    flex: none;
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    background: var(--color-surface-2);
    color: var(--color-text-muted);
    font-size: var(--fs-xs);
    font-weight: 600;
  }
  /* The icon covers the initials; a broken icon removes itself and uncovers them. */
  .tile img {
    position: absolute;
    inset: 0;
    box-sizing: border-box;
    width: 100%;
    height: 100%;
    padding: var(--space-2);
    border-radius: inherit;
    background: var(--color-surface-2);
  }
  .tags {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .shared {
    font-size: var(--fs-xs);
    opacity: 0.7;
  }
  /* Edit lies over the end of the row and shows on hover or focus, so the tag stays at the right edge. */
  .tools {
    position: absolute;
    inset: 50% 0 auto auto;
    transform: translateY(-50%);
    background: var(--color-surface);
    opacity: 0;
    pointer-events: none;
  }
  :global(.row:hover) .tools,
  .tools:focus-within {
    opacity: 1;
    pointer-events: auto;
  }
  @media (hover: none) {
    .tools {
      position: static;
      transform: none;
      opacity: 1;
      pointer-events: auto;
    }
  }
</style>
