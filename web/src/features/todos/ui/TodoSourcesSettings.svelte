<script lang="ts">
  import PlainHttpWarning from '../../../shared/ui/PlainHttpWarning.svelte';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { onDestroy } from 'svelte';
  import {
    type DiscoveredList,
    isOAuthKind,
    OAUTH_CONSOLES,
    SOURCE_KINDS,
    type SourceKind,
    type TodoSource,
  } from '../domain/todo';
  import type { TodoSourcesUseCases } from '../usecase/todosUseCases';

  /** Settings → To-do lists: connect CalDAV (#169), Todoist (#170) and OAuth lists (#171). */
  interface Props {
    sources: TodoSourcesUseCases;
  }
  let { sources }: Props = $props();

  let lists = $state<TodoSource[]>([]);
  let kind = $state<SourceKind>('caldav');
  let name = $state('');
  let url = $state('');
  let username = $state('');
  let password = $state('');
  let found = $state<DiscoveredList[] | null>(null);
  let chosen = $state('');
  let error = $state('');
  let message = $state('');
  let busy = $state(false);
  let redirect = $state('');
  let oauthState = $state('');
  let stopped = false;
  onDestroy(() => (stopped = true));
  const oauth = $derived(isOAuthKind(kind));
  const service = $derived(t(`todos.sources.kind_${kind}`));

  async function load() {
    lists = await sources.list(true);
  }
  $effect(() => {
    void load();
    void sources.oauthRedirect().then((r) => (redirect = r));
  });

  function reset() {
    found = null;
    chosen = oauthState = '';
  }

  /** Opens the provider's sign-in in a new window and waits for it (#171). */
  function signIn() {
    // Opened synchronously in the click, so pop-up blockers allow it.
    const win = window.open('about:blank', 'waylight-oauth', 'width=520,height=700');
    if (!win) {
      error = t('todos.sources.popup_blocked');
      return;
    }
    win.opener = null; // the provider's page gets no handle on this one
    void run(async () => {
      reset();
      try {
        const start = await sources.startOAuth(input());
        win.location.href = start.url;
        message = t('todos.sources.signing_in');
        const status = await sources.waitForSignIn(start.state, () => stopped);
        if (status !== 'ready') {
          message = '';
          if (status === 'failed') error = t('todos.sources.sign_in_failed');
          return;
        }
        oauthState = start.state;
        message = t('todos.sources.signed_in');
        found = await sources.discover({ ...input(), oauthState });
        chosen = found[0] ? value(found[0]) : '';
        if (found.length === 1 && !name) name = found[0]?.name ?? '';
        if (found.length === 0) error = t('todos.sources.none_found');
      } catch (err) {
        win.close();
        throw err;
      }
    });
  }

  async function run(action: () => Promise<void>) {
    error = message = '';
    busy = true;
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  const input = () => ({ kind, name, url, username, password });
  /** A found list's value: the CalDAV address or the Todoist project. */
  const value = (l: DiscoveredList) => l.listId || l.url;

  function find() {
    void run(async () => {
      found = await sources.discover(input());
      chosen = found[0] ? value(found[0]) : '';
      if (found.length === 1 && !name) name = found[0]?.name ?? '';
      if (found.length === 0) error = t('todos.sources.none_found');
    });
  }

  function connect(e: SubmitEvent) {
    e.preventDefault();
    void run(async () => {
      await sources.save(
        '',
        oauth
          ? { ...input(), url: '', listId: chosen, oauthState }
          : kind === 'todoist'
            ? { ...input(), url: '', username: '', listId: chosen }
            : { ...input(), url: chosen || url },
      );
      name = url = username = password = '';
      reset();
      message = t('todos.sources.saved');
      await load();
    });
  }

  function remove(s: TodoSource) {
    void run(async () => {
      await sources.delete(s.id);
      message = t('todos.sources.removed');
      await load();
    });
  }
</script>

<p class="muted">{t('todos.sources.intro')}</p>
<h3>{t('todos.sources.connected')}</h3>
<ul class="lists">
  {#each lists as s (s.id)}
    <li>
      <span class="what"
        ><strong>{s.name}</strong><span class="muted small"
          >{s.kind === 'caldav' ? `${s.username} · ${s.url}` : t(`todos.sources.kind_${s.kind}`)}</span
        ></span
      >
      <button
        class="btn btn-icon"
        type="button"
        disabled={busy}
        aria-label={t('todos.sources.remove', { name: s.name })}
        onclick={() => remove(s)}>✕</button
      >
    </li>
  {:else}
    <li class="muted">{t('todos.sources.none')}</li>
  {/each}
</ul>

<h3>{t('todos.sources.add')}</h3>
<form onsubmit={connect}>
  <label class="field"
    ><span>{t('todos.sources.kind')}</span>
    <select class="input" bind:value={kind} onchange={reset}>
      {#each SOURCE_KINDS as k (k)}<option value={k}>{t(`todos.sources.kind_${k}`)}</option>{/each}
    </select>
  </label>
  {#if kind === 'caldav'}
    <label class="field"
      ><span>{t('todos.sources.url')}</span><input
        class="input"
        type="url"
        bind:value={url}
        placeholder="https://cloud.example.com/remote.php/dav"
        required
      /></label
    >
    <p class="muted small">{t('todos.sources.url_hint')}</p>
    <PlainHttpWarning {url} />
    <label class="field"
      ><span>{t('todos.sources.username')}</span><input
        class="input"
        bind:value={username}
        autocomplete="off"
        required
      /></label
    >
  {/if}
  {#if oauth}
    {#if redirect}
      <p class="muted small">{t('todos.sources.oauth_intro', { service })}</p>
      <p><code class="redirect">{redirect}</code></p>
      {#if OAUTH_CONSOLES[kind]}<p class="small">
          <a href={OAUTH_CONSOLES[kind]} target="_blank" rel="noopener noreferrer"
            >{t('todos.sources.oauth_console', { service })}</a
          >
        </p>{/if}
      <label class="field"
        ><span>{t('todos.sources.client_id')}</span><input
          class="input"
          bind:value={username}
          autocomplete="off"
          required
        /></label
      >
      <label class="field"
        ><span>{t('todos.sources.client_secret')}</span><input
          class="input"
          type="password"
          bind:value={password}
          autocomplete="new-password"
        /></label
      >
      <p>
        <button class="btn" type="button" disabled={busy || !username} onclick={signIn}
          >{t('todos.sources.sign_in')}</button
        >
      </p>
    {:else}
      <p class="error">{t('todos.sources.oauth_unavailable')}</p>
    {/if}
  {:else}
    <label class="field"
      ><span>{kind === 'todoist' ? t('todos.sources.token') : t('todos.sources.password')}</span><input
        class="input"
        type="password"
        bind:value={password}
        autocomplete="new-password"
        required
      /></label
    >
    {#if kind === 'todoist'}<p class="muted small">{t('todos.sources.token_hint')}</p>{/if}
    <p><button class="btn" type="button" disabled={busy} onclick={find}>{t('todos.sources.find')}</button></p>
  {/if}
  {#if found && found.length > 0}
    <label class="field"
      ><span>{kind === 'caldav' ? t('todos.sources.found') : t('todos.sources.project')}</span>
      <select class="input" bind:value={chosen}>
        {#each found as l (value(l))}<option value={value(l)}>{l.name}</option>{/each}
      </select>
    </label>
  {/if}
  <label class="field"
    ><span>{t('todos.sources.name')}</span><input
      class="input"
      bind:value={name}
      maxlength="80"
      required
    /></label
  >
  <p class="muted small">{t('todos.sources.limits')}</p>
  {#if kind === 'todoist'}<p class="muted small">{t('todos.sources.limits_todoist')}</p>{/if}
  <p>
    <button class="btn btn-primary" type="submit" disabled={busy || (oauth && !oauthState)}
      >{t('todos.sources.connect')}</button
    >
  </p>
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  .lists {
    list-style: none;
    margin: 0 0 var(--space-3);
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .what {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .what .small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .small {
    font-size: var(--fs-xs);
  }
  .redirect {
    overflow-wrap: anywhere;
    user-select: all;
  }
  h3 {
    font-size: var(--fs-sm);
    margin: var(--space-3) 0 var(--space-2);
  }
</style>
