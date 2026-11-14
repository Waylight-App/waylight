<script lang="ts">
  import PlainHttpWarning from '../../../shared/ui/PlainHttpWarning.svelte';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { DiscoveredList, TodoSource } from '../domain/todo';
  import type { TodoSourcesUseCases } from '../usecase/todosUseCases';

  /** Settings → To-do lists: connect CalDAV task lists (#169). */
  interface Props {
    sources: TodoSourcesUseCases;
  }
  let { sources }: Props = $props();

  let lists = $state<TodoSource[]>([]);
  let name = $state('');
  let url = $state('');
  let username = $state('');
  let password = $state('');
  let found = $state<DiscoveredList[] | null>(null);
  let chosen = $state('');
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  async function load() {
    lists = await sources.list(true);
  }
  $effect(() => {
    void load();
  });

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

  const input = () => ({ name, url, username, password });

  function find() {
    void run(async () => {
      found = await sources.discover(input());
      chosen = found[0]?.url ?? '';
      if (found.length === 1 && !name) name = found[0]?.name ?? '';
      if (found.length === 0) error = t('todos.sources.none_found');
    });
  }

  function connect(e: SubmitEvent) {
    e.preventDefault();
    void run(async () => {
      await sources.save('', { ...input(), url: chosen || url });
      name = url = username = password = chosen = '';
      found = null;
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
        ><strong>{s.name}</strong><span class="muted small">{s.username} · {s.url}</span></span
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
  <label class="field"
    ><span>{t('todos.sources.password')}</span><input
      class="input"
      type="password"
      bind:value={password}
      autocomplete="new-password"
      required
    /></label
  >
  <p><button class="btn" type="button" disabled={busy} onclick={find}>{t('todos.sources.find')}</button></p>
  {#if found && found.length > 0}
    <label class="field"
      ><span>{t('todos.sources.found')}</span>
      <select class="input" bind:value={chosen}>
        {#each found as l (l.url)}<option value={l.url}>{l.name}</option>{/each}
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
  <p><button class="btn btn-primary" type="submit" disabled={busy}>{t('todos.sources.connect')}</button></p>
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
    font-size: 0.85em;
  }
  h3 {
    font-size: 1rem;
    margin: var(--space-3) 0 var(--space-2);
  }
</style>
