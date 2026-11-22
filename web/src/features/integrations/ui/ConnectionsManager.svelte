<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    type Connection,
    type ConnectionDraft,
    type ConnectionService,
    type ConnectionState,
    MAX_NAME_LENGTH,
    groupByService,
    isKnownField,
    stateCounts,
    usageList,
    validateDraft,
  } from '../domain/integrations';
  import type { ConnectionsUseCases } from '../usecase/connectionsUseCases';

  /** Props of the connections list in Settings → Connections & services (#708). */
  interface Props {
    connections: ConnectionsUseCases;
    /** Display name of a widget kind. */
    widgetName: (kind: string) => string;
    /** The account's tabs, to say where a widget is. */
    tabs: () => Promise<{ id: string; name: string }[]>;
  }
  let { connections, widgetName, tabs }: Props = $props();

  type Editing =
    | { kind: 'details'; id: string; name: string; url: string }
    | { kind: 'credential'; id: string; credential: Record<string, string> };

  let services = $state<ConnectionService[]>([]);
  let list = $state<Connection[]>([]);
  let states = $state<Record<string, ConnectionState>>({});
  let tabNames = $state<Map<string, string>>(new Map());
  let editing = $state<Editing | null>(null);
  let draft = $state<ConnectionDraft | null>(null);
  let error = $state('');
  let busy = $state(false);

  const groups = $derived(groupByService(list, services));
  const counts = $derived(stateCounts(Object.values(states)));
  const serviceOf = (id: string) => services.find((s) => s.service === id);
  const label = (field: string) => (isKnownField(field) ? t(`integrations.field.${field}`) : field);
  const uses = (c: Connection) =>
    usageList(c.usedBy, widgetName, (id) => tabNames.get(id), t('integrations.other_account'));

  async function check(c: Connection) {
    states[c.id] = { kind: 'checking' };
    states[c.id] = await connections.check(c.id);
  }

  async function load() {
    try {
      [services, list] = await Promise.all([connections.services(), connections.list()]);
      await Promise.all(list.map(check));
    } catch (err) {
      error = errorMessage(err);
    }
  }
  $effect(() => {
    void load();
    tabs()
      .then((all) => (tabNames = new Map(all.map((tab) => [tab.id, tab.name]))))
      .catch(() => undefined);
  });

  /** Runs a change, then reloads the list and checks the connection it touched. */
  async function run(change: () => Promise<Connection | undefined>, after?: () => void) {
    busy = true;
    error = '';
    try {
      const changed = await change();
      after?.();
      list = await connections.list();
      if (changed) await check(changed);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function startAdd(service: ConnectionService) {
    draft = { service: service.service, name: '', url: service.fixedUrl ?? '', credential: {} };
  }

  async function add(e: SubmitEvent) {
    e.preventDefault();
    if (!draft) return;
    const invalid = validateDraft(draft, serviceOf(draft.service));
    if (invalid) {
      error = t(invalid);
      return;
    }
    const d = draft;
    await run(
      () => connections.add(d),
      () => (draft = null),
    );
  }

  async function saveDetails(e: SubmitEvent, c: Connection) {
    e.preventDefault();
    if (editing?.kind !== 'details') return;
    const { name, url } = editing;
    await run(
      () => connections.edit(c, name, url),
      () => (editing = null),
    );
  }

  async function saveCredential(e: SubmitEvent, c: Connection) {
    e.preventDefault();
    if (editing?.kind !== 'credential') return;
    const { credential } = editing;
    await run(
      () => connections.replaceCredential(c.id, credential),
      () => (editing = null),
    );
  }

  async function remove(c: Connection) {
    const names = uses(c);
    const question = names.length
      ? t('integrations.confirm_remove_used', { name: c.name, list: names.join(', ') })
      : t('integrations.confirm_remove', { name: c.name });
    if (!confirm(question)) return;
    await run(() => connections.remove(c.id).then(() => undefined));
  }

  const stateText = (s: ConnectionState | undefined) =>
    !s || s.kind === 'unchecked'
      ? t('integrations.state_unchecked')
      : s.kind === 'checking'
        ? t('integrations.state_checking')
        : s.kind === 'ok'
          ? t('integrations.state_ok')
          : t('integrations.state_failing', { reason: s.message });
</script>

<h3>{t('integrations.title')}</h3>
<p class="muted">{t('integrations.intro')}</p>

{#if list.length > 0}
  <p class="muted" aria-live="polite">{t('integrations.summary', counts)}</p>
{:else}
  <p class="muted">{t('integrations.empty')}</p>
{/if}

{#each groups as group (group.service)}
  <h4>{group.title}</h4>
  <ul class="connections">
    {#each group.connections as c (c.id)}
      <li>
        {#if editing?.kind === 'details' && editing.id === c.id}
          <form class="row" onsubmit={(e) => saveDetails(e, c)}>
            <label class="field grow">
              <span>{t('integrations.name')}</span>
              <input class="input" bind:value={editing.name} maxlength={MAX_NAME_LENGTH} />
            </label>
            {#if !serviceOf(c.service)?.fixedUrl}
              <label class="field grow">
                <span>{t('integrations.url')}</span>
                <input class="input" bind:value={editing.url} />
              </label>
            {/if}
            <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
            <button class="btn" type="button" onclick={() => (editing = null)}
              >{t('integrations.cancel')}</button
            >
          </form>
        {:else if editing?.kind === 'credential' && editing.id === c.id}
          <form class="row" onsubmit={(e) => saveCredential(e, c)}>
            {#each serviceOf(c.service)?.fields ?? [] as f (f.name)}
              <label class="field grow">
                <span>{label(f.name)}</span>
                <input
                  class="input"
                  type={f.secret ? 'password' : 'text'}
                  autocomplete="off"
                  bind:value={editing.credential[f.name]}
                />
              </label>
            {/each}
            <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
            <button class="btn" type="button" onclick={() => (editing = null)}
              >{t('integrations.cancel')}</button
            >
          </form>
        {:else}
          <div class="text">
            <strong>{c.name}</strong>
            <span class="muted">{c.url}</span>
            <span class="muted small">{stateText(states[c.id])}</span>
            <span class="muted small"
              >{c.usedBy.length
                ? t('integrations.used_by', { list: uses(c).join(', ') })
                : t('integrations.unused')}</span
            >
          </div>
          <div class="actions">
            <button
              class="btn"
              type="button"
              aria-label={t('integrations.check_connection', { name: c.name })}
              onclick={() => check(c)}>{t('integrations.check')}</button
            >
            <button
              class="btn"
              type="button"
              aria-label={t('integrations.edit_connection', { name: c.name })}
              onclick={() => (editing = { kind: 'details', id: c.id, name: c.name, url: c.url })}
              >{t('integrations.edit')}</button
            >
            <button
              class="btn"
              type="button"
              aria-label={t('integrations.replace_connection', { name: c.name })}
              onclick={() => (editing = { kind: 'credential', id: c.id, credential: {} })}
              >{t('integrations.replace')}</button
            >
            <button
              class="btn btn-danger"
              type="button"
              aria-label={t('integrations.remove_connection', { name: c.name })}
              onclick={() => remove(c)}>{t('integrations.remove')}</button
            >
          </div>
        {/if}
      </li>
    {/each}
  </ul>
{/each}

<h4>{t('integrations.add')}</h4>
{#if draft}
  <form class="row" onsubmit={add}>
    <label class="field grow">
      <span>{t('integrations.name')}</span>
      <input class="input" bind:value={draft.name} maxlength={MAX_NAME_LENGTH} />
    </label>
    {#if !serviceOf(draft.service)?.fixedUrl}
      <label class="field grow">
        <span>{t('integrations.url')}</span>
        <input class="input" bind:value={draft.url} placeholder="https://" />
      </label>
    {/if}
    {#each serviceOf(draft.service)?.fields ?? [] as f (f.name)}
      <label class="field grow">
        <span>{label(f.name)}{f.optional ? ` (${t('integrations.optional')})` : ''}</span>
        <input
          class="input"
          type={f.secret ? 'password' : 'text'}
          autocomplete="off"
          bind:value={draft.credential[f.name]}
        />
      </label>
    {/each}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('integrations.save')}</button>
    <button class="btn" type="button" onclick={() => (draft = null)}>{t('integrations.cancel')}</button>
  </form>
{:else}
  <label class="field">
    <span>{t('integrations.service')}</span>
    <select
      class="input"
      onchange={(e) => {
        const s = serviceOf(e.currentTarget.value);
        if (s) startAdd(s);
        e.currentTarget.value = '';
      }}
    >
      <option value="">{t('integrations.choose_service')}</option>
      {#each services as s (s.service)}
        <option value={s.service}>{s.title}</option>
      {/each}
    </select>
  </label>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .connections {
    list-style: none;
    margin: 0 0 var(--space-3);
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .connections li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .small {
    font-size: 0.85rem;
  }
  .actions,
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: end;
    gap: var(--space-2);
  }
  .row {
    margin-bottom: var(--space-2);
  }
  .grow {
    flex: 1;
    min-width: 12rem;
  }
</style>
