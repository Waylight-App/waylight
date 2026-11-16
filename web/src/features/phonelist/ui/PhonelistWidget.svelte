<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { emptyInput, filterEntries, type PhoneEntry, type PhoneEntryInput } from '../domain/phonelist';
  import type { PhonelistUseCases } from '../usecase/phonelistUseCases';

  /** Props of the phone list widget (#300). */
  interface Props {
    phonelist: PhonelistUseCases;
    /** Admins maintain the list; everyone else reads it. */
    admin: boolean;
  }
  let { phonelist, admin }: Props = $props();

  let entries = $state<PhoneEntry[]>([]);
  let query = $state('');
  let error = $state('');
  // The entry being edited, 'new' for the add form, or null.
  let editing = $state<PhoneEntry | 'new' | null>(null);
  let form = $state<PhoneEntryInput>(emptyInput());

  const shown = $derived(filterEntries(entries, query));

  const load = () =>
    phonelist
      .list()
      .then((l) => ((entries = l), (error = '')))
      .catch((err) => (error = errorMessage(err)));

  $effect(() => {
    load();
  });

  function edit(target: PhoneEntry | 'new') {
    editing = target;
    form = target === 'new' ? emptyInput() : { ...target };
    error = '';
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    try {
      await phonelist.save(form, editing === 'new' || editing === null ? undefined : editing.id);
      editing = null;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(entry: PhoneEntry) {
    if (!confirm(t('phonelist.confirm_delete', { name: entry.name }))) return;
    try {
      await phonelist.remove(entry.id);
      editing = null;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<div class="bar">
  <input
    class="input"
    type="search"
    bind:value={query}
    placeholder={t('phonelist.search')}
    aria-label={t('phonelist.search')}
  />
  {#if admin}
    <button class="btn btn-icon" type="button" aria-label={t('phonelist.add')} onclick={() => edit('new')}
      >＋</button
    >
  {/if}
</div>
{#if error}<p class="error" role="alert">{error}</p>{/if}

{#if editing}
  <form class="edit" onsubmit={save}>
    <input
      class="input"
      bind:value={form.name}
      placeholder={t('phonelist.name')}
      aria-label={t('phonelist.name')}
      required
    />
    <input
      class="input"
      bind:value={form.role}
      placeholder={t('phonelist.role')}
      aria-label={t('phonelist.role')}
    />
    <input
      class="input"
      bind:value={form.team}
      placeholder={t('phonelist.team')}
      aria-label={t('phonelist.team')}
    />
    <input
      class="input"
      type="tel"
      bind:value={form.phone}
      placeholder={t('phonelist.phone')}
      aria-label={t('phonelist.phone')}
    />
    <input
      class="input"
      type="email"
      bind:value={form.email}
      placeholder={t('phonelist.email')}
      aria-label={t('phonelist.email')}
    />
    <div class="actions">
      <button class="btn btn-primary" type="submit">{t('common.save')}</button>
      <button class="btn" type="button" onclick={() => (editing = null)}>{t('common.cancel')}</button>
      {#if editing !== 'new'}
        {@const target = editing}
        <button class="btn btn-danger" type="button" onclick={() => remove(target)}
          >{t('phonelist.delete')}</button
        >
      {/if}
    </div>
  </form>
{/if}

<ul>
  {#each shown as e (e.id)}
    <li>
      <div class="who">
        <strong>{e.name}</strong>
        {#if e.role || e.team}<span class="muted">{[e.role, e.team].filter(Boolean).join(' · ')}</span>{/if}
      </div>
      <div class="how">
        {#if e.dial}<a href="tel:{e.dial}">{e.phone}</a>{/if}
        {#if e.email}<a href="mailto:{e.email}">{e.email}</a>{/if}
      </div>
      {#if e.readOnly}
        <span class="book" title={t('phonelist.from_book')} aria-label={t('phonelist.from_book')} role="img"
          >📇</span
        >
      {:else if admin}
        <button
          class="btn btn-icon edit-btn"
          type="button"
          aria-label={t('phonelist.edit', { name: e.name })}
          onclick={() => edit(e)}>✎</button
        >
      {/if}
    </li>
  {:else}
    <li class="muted">
      {query ? t('phonelist.none_match') : admin ? t('phonelist.empty_admin') : t('phonelist.empty')}
    </li>
  {/each}
</ul>

<style>
  .bar {
    display: flex;
    gap: var(--space-1);
    margin-bottom: var(--space-2);
  }
  .bar .input {
    flex: 1;
    min-width: 0;
  }
  .edit {
    display: grid;
    gap: var(--space-1);
    margin-bottom: var(--space-2);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0 var(--space-2);
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  .who {
    flex: 1 1 10rem;
    display: flex;
    flex-direction: column;
  }
  .who .muted {
    font-size: 0.85em;
  }
  .how {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    font-size: 0.9em;
  }
  .edit-btn {
    opacity: 0;
  }
  .book {
    font-size: 0.8em;
    opacity: 0.6;
  }
  li:hover .edit-btn,
  .edit-btn:focus {
    opacity: 1;
  }
</style>
