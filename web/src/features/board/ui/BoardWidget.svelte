<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import type { BoardNote } from '../domain/board';
  import type { BoardUseCases } from '../usecase/boardUseCases';
  import NoteDialog from './NoteDialog.svelte';

  /** Props of the family board widget (#685). */
  interface Props {
    board: BoardUseCases;
  }
  let { board }: Props = $props();
  const poll = usePolling();

  let notes = $state<BoardNote[]>([]);
  let error = $state('');
  let editing = $state<BoardNote | null | undefined>(undefined); // null = new

  const todayIso = () => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  };

  async function load() {
    try {
      notes = await board.list();
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  $effect(() => poll(load, 5 * 60_000));

  function until(day: string): string {
    const date = new Date(`${day}T12:00:00`).toLocaleDateString(intlLocale(), {
      day: 'numeric',
      month: 'short',
    });
    return t('board.until', { date });
  }

  async function pin(n: BoardNote) {
    try {
      await board.setPinned(n, !n.pinned);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(n: BoardNote) {
    if (!confirm(t('board.confirm_remove'))) return;
    try {
      await board.remove(n.id);
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if error}<p class="error" role="alert">{error}</p>{/if}
<ul>
  {#each notes as n (n.id)}
    <ListRow>
      <p class="text">{n.text}</p>
      {#snippet sub()}
        {#if n.pinned}<Icon name="pin" size={16} /> {t('board.pinned')}{/if}
        {#if n.expiresOn}{n.pinned ? ' · ' : ''}{until(n.expiresOn)}{/if}
      {/snippet}
      {#snippet trail()}
        {#if n.mine}
          <span class="actions">
            <button
              class="btn btn-icon"
              type="button"
              aria-pressed={n.pinned}
              aria-label={n.pinned ? t('board.unpin') : t('board.pin')}
              onclick={() => pin(n)}><Icon name="pin" size={16} /></button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('board.edit')}
              onclick={() => (editing = n)}>✎</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('board.remove')}
              onclick={() => remove(n)}>✕</button
            >
          </span>
        {/if}
      {/snippet}
    </ListRow>
  {:else}
    <li class="muted empty">{t('board.empty')}</li>
  {/each}
</ul>
<button class="btn add" type="button" onclick={() => (editing = null)}>+ {t('board.new')}</button>

{#if editing !== undefined}
  <NoteDialog
    {board}
    note={editing}
    today={todayIso()}
    onsaved={() => load().catch(() => {})}
    onclose={() => (editing = undefined)}
  />
{/if}

<style>
  ul {
    list-style: none;
    margin: 0 0 var(--space-2);
    padding: 0;
  }
  .text {
    margin: 0;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }
  .actions {
    display: flex;
  }
</style>
