<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { ConnectionsApi, outcome, type Clients, type Connection } from './connections';
  import { texts } from './texts';

  /** Settings → Connected accounts (#326): every member's own links, and the setup for admins. */
  interface Props {
    context: PluginContext;
  }
  let { context }: Props = $props();

  const api = $derived(new ConnectionsApi(context.http));
  const tr = (key: string, params: Record<string, string | number> = {}) =>
    text(texts, key, context.language, params);
  const readOnly = $derived(context.level('connected-accounts') === 'read');

  let list = $state<Connection[]>([]);
  let clients = $state<Clients | null>(null);
  let drafts = $state<Record<string, { clientId: string; clientSecret: string; tenant: string }>>({});
  let error = $state('');
  let message = $state('');

  const nameOf = (service: string) => list.find((c) => c.service === service)?.name ?? service;

  async function load() {
    try {
      list = await api.list();
      if (context.account.admin) {
        clients = await api.clients();
        drafts = Object.fromEntries(
          clients.clients.map((c) => [
            c.service,
            { clientId: c.clientId, clientSecret: '', tenant: c.tenant },
          ]),
        );
      }
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  $effect(() => {
    load().then(() => {
      // Back from the service: say how it went once, then tidy the address bar.
      const o = outcome(location.search);
      if (o.connected) message = tr('done', { name: nameOf(o.connected) });
      if (o.failed) error = tr('failed', { name: nameOf(o.failed) });
      if (o.connected || o.failed) history.replaceState(null, '', location.pathname);
    });
  });

  async function connect(c: Connection) {
    error = message = '';
    try {
      // The service's sign-in page; it sends the browser back to the settings.
      location.assign(await api.start(c.service));
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function disconnect(c: Connection) {
    error = message = '';
    try {
      await api.disconnect(c.service);
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function saveClient(e: SubmitEvent, service: string) {
    e.preventDefault();
    error = message = '';
    const d = drafts[service];
    if (!d) return;
    try {
      await api.saveClient(service, d);
      message = tr('saved');
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }
</script>

<p class="muted">{tr('description')}</p>
<ul>
  {#each list as c (c.service)}
    <li>
      <div>
        <strong>{c.name}</strong>
        <span class="muted small">
          {#if !c.configured}{tr('not_configured')}
          {:else if c.status === 'connected' && c.since}{tr('connected_since', {
              date: new Date(c.since).toLocaleDateString(context.language),
            })}
          {:else if c.status === 'expired'}{tr('expired')}
          {:else}{tr('not_connected')}{/if}
        </span>
      </div>
      {#if c.configured && !readOnly}
        {#if c.status === 'connected'}
          <button class="btn" type="button" onclick={() => disconnect(c)}>{tr('disconnect')}</button>
        {:else}
          <button class="btn btn-primary" type="button" onclick={() => connect(c)}
            >{c.status === 'expired' ? tr('reconnect') : tr('connect')}</button
          >
        {/if}
      {/if}
    </li>
  {/each}
</ul>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

{#if clients && !readOnly}
  <h3>{tr('admin_title')}</h3>
  <p class="muted small">{tr('admin_hint')} <code>{clients.redirectUrl}</code></p>
  {#each clients.clients as c (c.service)}
    {@const d = drafts[c.service]}
    {#if d}
      <form onsubmit={(e) => saveClient(e, c.service)}>
        <strong>{c.name}</strong>
        <label class="field"
          ><span>{tr('client_id')}</span><input
            class="input"
            bind:value={d.clientId}
            autocomplete="off"
          /></label
        >
        <label class="field"
          ><span>{tr('client_secret')}</span><input
            class="input"
            type="password"
            autocomplete="off"
            bind:value={d.clientSecret}
            placeholder={c.hasSecret ? tr('secret_stored') : ''}
          /></label
        >
        {#if c.service === 'microsoft'}
          <label class="field"
            ><span>{tr('tenant')}</span><input
              class="input"
              bind:value={d.tenant}
              autocomplete="off"
            /></label
          >
        {/if}
        <button class="btn" type="submit">{tr('save')}</button>
      </form>
    {/if}
  {/each}
{/if}

<style>
  ul {
    list-style: none;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  li div {
    flex: 1;
    display: flex;
    flex-direction: column;
  }
  .small {
    font-size: var(--fs-xs);
  }
  h3 {
    font-size: var(--fs-sm);
    margin: var(--space-3) 0 var(--space-1);
  }
  form {
    display: grid;
    gap: var(--space-2);
    margin: var(--space-2) 0 var(--space-3);
  }
  code {
    overflow-wrap: anywhere;
  }
</style>
