<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import {
    AbsencesApi,
    addDays,
    dayKey,
    enterableKinds,
    sections,
    validate,
    type Absence,
    type AbsenceInput,
  } from './absences';
  import { texts } from './texts';

  /** The absences widget (#299): who is away today and this week. */
  interface Props {
    context: PluginContext;
    config?: Record<string, unknown>;
  }
  let { context }: Props = $props();

  const api = $derived(new AbsencesApi(context.http));
  const tr = (key: string, params: Record<string, string | number> = {}) =>
    text(texts, key, context.language, params);
  const date = (day: string) =>
    new Date(`${day}T00:00:00`).toLocaleDateString(context.language, {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
    });

  // Entering needs a key that is not expired; guests and kiosk displays only read.
  const readOnly = $derived(context.level('absences') === 'read');
  const mayEnter = $derived(!context.account.guest && !readOnly);

  let list = $state<Absence[]>([]);
  let members = $state<{ id: string; name: string }[]>([]);
  let editing = $state<Absence | 'new' | null>(null);
  let form = $state<AbsenceInput>(blank());
  let error = $state('');
  let today = $state(dayKey(new Date()));
  const parts = $derived(sections(list, today));

  function blank(): AbsenceInput {
    const d = dayKey(new Date());
    return { accountId: '', from: d, to: d, kind: 'holiday', note: '', private: false };
  }

  async function load() {
    today = dayKey(new Date());
    try {
      list = await api.list(today, addDays(today, 6));
      error = '';
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  $effect(() => {
    load();
    const timer = setInterval(load, 15 * 60_000);
    return () => clearInterval(timer);
  });

  $effect(() => {
    if (!context.account.admin) return;
    api
      .members()
      .then((m) => (members = m))
      .catch(() => (members = []));
  });

  function startEdit(a: Absence | 'new') {
    editing = a;
    error = '';
    form =
      a === 'new'
        ? blank()
        : {
            accountId: a.accountId,
            from: a.from,
            to: a.to,
            kind: a.kind === 'away' ? 'other' : a.kind,
            note: a.note,
            private: a.private,
          };
  }

  async function save(e: SubmitEvent) {
    e.preventDefault();
    const problem = validate(form);
    if (problem) {
      error = tr(`error_${problem}`);
      return;
    }
    try {
      await api.save(form, editing && editing !== 'new' ? editing.id : undefined);
      editing = null;
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function remove(a: Absence) {
    if (!confirm(tr('delete_confirm'))) return;
    try {
      await api.remove(a.id);
      editing = null;
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }
</script>

{#snippet row(a: Absence, showStart: boolean)}
  <li>
    <span class="who">
      <strong>{a.name}</strong>
      <span class="muted small">
        {tr(`kind_${a.kind}`)}{#if a.private && a.kind !== 'away'}
          · {tr('private_badge')}{/if}
        · {showStart && a.from !== a.to
          ? tr('range', { from: date(a.from), to: date(a.to) })
          : showStart
            ? date(a.from)
            : tr('until', { date: date(a.to) })}
      </span>
      {#if a.note}<span class="small note">{a.note}</span>{/if}
    </span>
    {#if a.editable && mayEnter}
      <button
        class="btn btn-icon"
        type="button"
        aria-label={tr('edit', { name: a.name })}
        onclick={() => startEdit(a)}>✎</button
      >
    {/if}
  </li>
{/snippet}

<div class="bar">
  {#if mayEnter}
    <button class="btn" type="button" onclick={() => startEdit('new')}>＋ {tr('add')}</button>
  {/if}
</div>
{#if readOnly}<p class="muted small">{tr('read_only')}</p>{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

{#if editing}
  <form class="edit" onsubmit={save}>
    {#if editing === 'new' && context.account.admin && members.length}
      <label class="field"
        ><span>{tr('field_person')}</span><select class="input" bind:value={form.accountId}>
          <option value="">{tr('me')}</option>
          {#each members as m (m.id)}<option value={m.id}>{m.name}</option>{/each}
        </select></label
      >
    {/if}
    <div class="row">
      <label class="field"
        ><span>{tr('field_from')}</span><input
          class="input"
          type="date"
          bind:value={form.from}
          required
        /></label
      >
      <label class="field"
        ><span>{tr('field_to')}</span><input
          class="input"
          type="date"
          bind:value={form.to}
          min={form.from}
          required
        /></label
      >
    </div>
    <label class="field"
      ><span>{tr('field_kind')}</span><select class="input" bind:value={form.kind}>
        {#each enterableKinds as k (k)}<option value={k}>{tr(`kind_${k}`)}</option>{/each}
      </select></label
    >
    <label class="field"
      ><span>{tr('field_note')}</span><input class="input" bind:value={form.note} maxlength="200" /></label
    >
    <label class="check"><input type="checkbox" bind:checked={form.private} /> {tr('field_private')}</label>
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

<h4>{tr('today')}</h4>
<ul>
  {#each parts.today as a (a.id)}{@render row(a, false)}{:else}<li class="muted">
      {tr('nobody_today')}
    </li>{/each}
</ul>
<h4>{tr('week')}</h4>
<ul>
  {#each parts.week as a (a.id)}{@render row(a, true)}{:else}<li class="muted">{tr('nobody_week')}</li>{/each}
</ul>

<style>
  .bar {
    display: flex;
    justify-content: flex-end;
    margin-bottom: var(--space-1);
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: var(--fs-sm);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  li {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  .who {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .note {
    overflow-wrap: anywhere;
  }
  .small {
    font-size: var(--fs-xs);
  }
  .edit {
    display: grid;
    gap: var(--space-1);
    margin-bottom: var(--space-2);
  }
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-1);
  }
  .field {
    display: grid;
    gap: 2px;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
  }
</style>
