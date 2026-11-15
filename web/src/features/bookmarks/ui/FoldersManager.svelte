<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Folder } from '../domain/bookmark';
  import type { BookmarksUseCases } from '../usecase/bookmarksUseCases';
  import type { ShareGroup } from '../../../shared/sharing/sharing';
  import VisibilitySelect from '../../../shared/ui/VisibilitySelect.svelte';

  /** Props of the bookmark folder management section. */
  interface Props {
    bookmarks: BookmarksUseCases;
  }
  let { bookmarks }: Props = $props();

  let folders = $state<Folder[]>([]);
  let name = $state('');
  let error = $state('');
  let groups = $state<ShareGroup[]>([]);

  async function run(action: () => Promise<unknown>) {
    error = '';
    try {
      await action();
      folders = (await bookmarks.list()).folders;
    } catch (err) {
      error = errorMessage(err);
    }
  }

  run(async () => {});
  $effect(() => {
    bookmarks
      .shareGroups()
      .then((g) => (groups = g))
      .catch(() => (groups = []));
  });
</script>

<p class="muted">{t('bookmarks.bookmarks_are_added_and_edited_directly')}</p>
<p class="muted">{t('bookmarks.folder_sharing_hint')}</p>
<ul class="list">
  {#each folders as f (f.id)}
    <li>
      <span class="name">{f.name} <span class="muted">({f.count})</span></span>
      {#if !f.mine}
        <span class="muted small">{t('bookmarks.shared_with_you')}</span>
      {:else}
        <VisibilitySelect
          value={f.visibility}
          {groups}
          onchange={(v) => run(() => bookmarks.shareFolder(f.id, v))}
        />
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('bookmarks.rename', { name: f.name })}
          onclick={() => {
            const n = prompt(t('bookmarks.folder_name'), f.name);
            if (n) run(() => bookmarks.renameFolder(f.id, n));
          }}>✎</button
        >
        <button
          class="btn btn-icon"
          type="button"
          aria-label={t('bookmarks.delete', { name: f.name })}
          onclick={() =>
            confirm(t('bookmarks.confirm_delete_folder', { name: f.name })) &&
            run(() => bookmarks.deleteFolder(f.id))}>✕</button
        >
      {/if}
    </li>
  {/each}
</ul>
<form
  class="add"
  onsubmit={(e) => {
    e.preventDefault();
    run(() => bookmarks.createFolder(name)).then(() => (name = ''));
  }}
>
  <input
    class="input"
    bind:value={name}
    placeholder={t('bookmarks.new_folder')}
    required
    maxlength="60"
    aria-label={t('bookmarks.folder_name')}
  />
  <button class="btn" type="submit">{t('bookmarks.add')}</button>
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .list {
    list-style: none;
    padding: 0;
  }
  .list li {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
    padding: var(--space-1) 0;
  }
  .small {
    font-size: 0.85em;
  }
  .name {
    flex: 1;
  }
  .add {
    display: flex;
    gap: var(--space-2);
  }
</style>
