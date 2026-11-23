<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { renderMarkdown } from '../lib/markdown';
  import { text } from '../lib/texts';
  import {
    AnnouncementsApi,
    dayOf,
    unreadCount,
    type Announcement,
    type AnnouncementInput,
  } from './announcements';
  import { texts } from './texts';

  /** The announcements widget (#297): notices for everyone, written by admins. */
  interface Props {
    context: PluginContext;
    config?: Record<string, unknown>;
  }
  let { context }: Props = $props();

  const api = $derived(new AnnouncementsApi(context.http));
  const tr = (key: string, params: Record<string, string | number> = {}) =>
    text(texts, key, context.language, params);
  const date = (iso: string) => new Date(iso).toLocaleDateString(context.language);

  // Writing needs the server's yes (admins and the writer group) and a key
  // that is not expired.
  const readOnly = $derived(context.level('announcements') === 'read');
  let allowed = $state(false);
  const mayWrite = $derived(allowed && !readOnly);
  $effect(() => {
    if (context.account.guest) return;
    api
      .mayWrite()
      .then((ok) => (allowed = ok))
      .catch(() => (allowed = false));
  });
  // Guests and kiosk displays cannot keep read marks (the server answers 403).
  const mayMark = $derived(!context.account.guest);

  let list = $state<Announcement[]>([]);
  let archive = $state(false);
  let open = $state<string | null>(null);
  let editing = $state<Announcement | 'new' | null>(null);
  let form = $state<AnnouncementInput>({ title: '', body: '', pinned: false, expiresOn: '', notify: false });
  let error = $state('');

  async function load() {
    try {
      list = await api.list(archive);
      error = '';
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  $effect(() => {
    void archive;
    load();
  });

  function startEdit(a: Announcement | 'new') {
    editing = a;
    form =
      a === 'new'
        ? { title: '', body: '', pinned: false, expiresOn: '', notify: false }
        : { title: a.title, body: a.body, pinned: a.pinned, expiresOn: dayOf(a.expiresAt), notify: false };
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    try {
      await api.save(form, editing && editing !== 'new' ? editing.id : undefined);
      editing = null;
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function remove(a: Announcement) {
    if (!confirm(tr('delete_confirm', { title: a.title }))) return;
    try {
      await api.remove(a.id);
      editing = null;
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function markRead(a: Announcement) {
    try {
      await api.markRead(a.id);
      a.read = true;
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  function toggle(a: Announcement) {
    open = open === a.id ? null : a.id;
  }
</script>

<div class="bar">
  <button class="link" type="button" onclick={() => (archive = !archive)}
    >{archive ? tr('current') : tr('archive')}</button
  >
  {#if !archive && unreadCount(list)}<span class="count">{tr('unread', { n: unreadCount(list) })}</span>{/if}
  {#if mayWrite && !archive}
    <button
      class="btn btn-icon add"
      type="button"
      aria-label={tr('write')}
      title={tr('write')}
      onclick={() => startEdit('new')}>＋</button
    >
  {/if}
</div>
{#if readOnly}<p class="muted small">{tr('read_only')}</p>{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

{#if editing}
  <form class="edit" onsubmit={save}>
    <input
      class="input"
      bind:value={form.title}
      placeholder={tr('field_title')}
      aria-label={tr('field_title')}
      required
      maxlength="200"
    />
    <textarea
      class="input"
      rows="5"
      bind:value={form.body}
      placeholder={tr('field_body')}
      aria-label={tr('field_body')}></textarea>
    <label class="check"><input type="checkbox" bind:checked={form.pinned} /> {tr('field_pinned')}</label>
    <label class="field"
      ><span>{tr('field_expires')}</span><input
        class="input"
        type="date"
        bind:value={form.expiresOn}
      /></label
    >
    {#if editing === 'new'}
      <label class="check"><input type="checkbox" bind:checked={form.notify} /> {tr('field_notify')}</label>
    {/if}
    <div class="actions">
      <button class="btn btn-primary" type="submit">{tr('save')}</button>
      <button class="btn" type="button" onclick={() => (editing = null)}>{tr('cancel')}</button>
      {#if editing !== 'new'}
        {@const target = editing}
        <button class="btn btn-danger" type="button" onclick={() => remove(target)}>{tr('delete')}</button>
      {/if}
    </div>
  </form>
{/if}

<ul>
  {#each list as a (a.id)}
    <li class:unread={!a.read && !archive} class:pinned={a.pinned}>
      <button class="head" type="button" aria-expanded={open === a.id} onclick={() => toggle(a)}>
        {#if a.pinned}<span class="pin" title={tr('pinned')} aria-label={tr('pinned')} role="img">📌</span
          >{/if}
        <span class="t">{a.title}</span>
      </button>
      <div class="meta muted">
        {tr('by', { name: a.authorName, date: date(a.createdAt) })}
        {#if a.expiresAt}· {tr('until', { date: date(a.expiresAt) })}{/if}
      </div>
      {#if open === a.id}
        <!-- Sanitised by DOMPurify in renderMarkdown, as in the core. -->
        <div class="body">{@html renderMarkdown(a.body)}</div>
        <div class="actions">
          {#if mayMark && !a.read && !archive}
            <button class="btn" type="button" onclick={() => markRead(a)}>{tr('mark_read')}</button>
          {/if}
          {#if a.editable && !readOnly}
            <button class="btn" type="button" onclick={() => startEdit(a)}>{tr('edit')}</button>
          {/if}
        </div>
      {/if}
    </li>
  {:else}
    <li class="muted">{archive ? tr('archive_empty') : tr('empty')}</li>
  {/each}
</ul>

<style>
  .bar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .add {
    margin-left: auto;
  }
  .link {
    background: none;
    border: 0;
    padding: 0;
    color: var(--color-accent);
    cursor: pointer;
    font: inherit;
  }
  .count {
    font-size: var(--fs-xs);
    padding: 0 var(--space-2);
    border-radius: 999px;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
  }
  .small {
    font-size: var(--fs-xs);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  li.unread .t {
    font-weight: 700;
  }
  li.unread .head::before {
    content: '';
    display: inline-block;
    width: 0.5em;
    height: 0.5em;
    margin-right: var(--space-1);
    border-radius: 50%;
    background: var(--color-accent);
  }
  .head {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    width: 100%;
    background: none;
    border: 0;
    padding: 0;
    text-align: left;
    font: inherit;
    color: inherit;
    cursor: pointer;
  }
  .meta {
    font-size: var(--fs-xs);
  }
  .body {
    margin: var(--space-2) 0;
    overflow-wrap: anywhere;
  }
  .edit {
    display: grid;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
</style>
