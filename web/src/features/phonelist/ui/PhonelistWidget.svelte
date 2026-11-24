<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { copyToClipboard } from '../../../shared/ui/clipboard';
  import {
    emptyInput,
    filterEntries,
    groupByTeam,
    parsePhonelistConfig,
    type PhoneEntry,
    type PhoneEntryInput,
    toCsv,
  } from '../domain/phonelist';
  import type { PhonelistUseCases } from '../usecase/phonelistUseCases';

  /** Props of the phone list widget (#300). */
  interface Props {
    phonelist: PhonelistUseCases;
    /** Admins maintain the list; everyone else reads it. */
    admin: boolean;
    config?: Record<string, unknown>;
  }
  let { phonelist, admin, config = {} }: Props = $props();

  let entries = $state<PhoneEntry[]>([]);
  let query = $state('');
  let error = $state('');
  // The entry being edited, 'new' for the add form, or null.
  let editing = $state<PhoneEntry | 'new' | null>(null);
  let form = $state<PhoneEntryInput>(emptyInput());

  const shown = $derived(filterEntries(entries, query));
  const grouped = $derived(parsePhonelistConfig(config).groupByTeam);
  // An empty list keeps one group, so the empty or "nobody matches" text shows.
  const groups = $derived(grouped && shown.length > 0 ? groupByTeam(shown) : [{ team: '', entries: shown }]);
  // ID of the entry whose number was just copied (#416).
  let copied = $state('');

  async function copy(e: PhoneEntry) {
    if (await copyToClipboard(e.phone)) {
      copied = e.id;
      setTimeout(() => {
        if (copied === e.id) copied = '';
      }, 1500);
    }
  }

  function downloadCsv() {
    const header = [
      t('phonelist.name'),
      t('phonelist.role'),
      t('phonelist.team'),
      t('phonelist.phone'),
      t('phonelist.email'),
    ];
    const blob = new Blob([toCsv(entries, header)], { type: 'text/csv;charset=utf-8' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'phone-list.csv';
    a.click();
    URL.revokeObjectURL(a.href);
  }

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
    <button
      class="btn btn-icon"
      type="button"
      aria-label={t('phonelist.download_csv')}
      title={t('phonelist.download_csv')}
      disabled={entries.length === 0}
      onclick={downloadCsv}>⤓</button
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

{#each groups as g (g.team)}
  {#if grouped && g.entries.length > 0}<h4 class="team">{g.team || t('phonelist.other_team')}</h4>{/if}
  <ul>
    {#each g.entries as e (e.id)}
      <ListRow>
        <strong>{e.name}</strong>
        {#snippet sub()}
          {#if e.role || (!grouped && e.team)}{[e.role, grouped ? '' : e.team]
              .filter(Boolean)
              .join(' · ')}{/if}
        {/snippet}
        {#snippet value()}
          <span class="how">
            {#if e.dial}
              <span class="number">
                <a href="tel:{e.dial}">{e.phone}</a>
                <button
                  class="btn btn-icon copy"
                  type="button"
                  aria-label={copied === e.id
                    ? t('phonelist.copied')
                    : t('phonelist.copy', { phone: e.phone })}
                  title={copied === e.id ? t('phonelist.copied') : t('phonelist.copy', { phone: e.phone })}
                  onclick={() => copy(e)}
                  >{#if copied === e.id}<Icon name="done" size={16} />{:else}<Icon
                      name="copy"
                      size={16}
                    />{/if}</button
                >
              </span>
            {/if}
            {#if e.email}<a href="mailto:{e.email}">{e.email}</a>{/if}
          </span>
        {/snippet}
        {#snippet trail()}
          {#if e.readOnly}
            <span
              class="book"
              title={t('phonelist.from_book')}
              aria-label={t('phonelist.from_book')}
              role="img"><Icon name="user" size={16} /></span
            >
          {:else if admin}
            <button
              class="btn btn-icon edit-btn"
              type="button"
              aria-label={t('phonelist.edit', { name: e.name })}
              onclick={() => edit(e)}>✎</button
            >
          {/if}
        {/snippet}
      </ListRow>
    {:else}
      <li class="muted">
        {query ? t('phonelist.none_match') : admin ? t('phonelist.empty_admin') : t('phonelist.empty')}
      </li>
    {/each}
  </ul>
{/each}

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
  .how {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }
  .edit-btn {
    opacity: 0;
  }
  .number {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
  }
  .copy {
    font-size: var(--fs-xs);
    padding: 0 var(--space-1);
  }
  .team {
    margin: var(--space-2) 0 0;
    font-size: var(--fs-xs);
  }
  .book {
    font-size: var(--fs-xs);
    opacity: 0.6;
  }
  :global(.row:hover) .edit-btn,
  .edit-btn:focus {
    opacity: 1;
  }
</style>
