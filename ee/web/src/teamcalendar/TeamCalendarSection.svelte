<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { TeamCalendarApi, type Group, type Subscription } from './teamcalendar';
  import { texts } from './texts';

  /** Settings → Team calendar (#299): share own calendar subscriptions with a group. */
  interface Props {
    context: PluginContext;
  }
  let { context }: Props = $props();

  const api = $derived(new TeamCalendarApi(context.http));
  const tr = (key: string, params: Record<string, string | number> = {}) =>
    text(texts, key, context.language, params);
  const readOnly = $derived(context.level('absences') === 'read');

  let list = $state<Subscription[]>([]);
  let accounts = $state<Subscription[]>([]);
  let groups = $state<Group[]>([]);
  let error = $state('');
  let message = $state('');

  $effect(() => {
    Promise.all([api.subscriptions(), api.accounts(), api.groups()])
      .then(([s, a, g]) => ((list = s), (accounts = a), (groups = g)))
      .catch((err) => (error = context.errorMessage(err)));
  });

  async function share(s: Subscription, visibility: string, account = false) {
    error = message = '';
    try {
      s.visibility = (
        await (account ? api.shareAccount(s.id, visibility) : api.share(s.id, visibility))
      ).visibility;
      message = tr('saved');
    } catch (err) {
      error = context.errorMessage(err);
    }
  }
</script>

<p class="muted">{tr('description')}</p>
<p class="muted small">{tr('hint')}</p>
{#if readOnly}<p class="muted">{tr('read_only')}</p>{/if}
{#snippet row(s: Subscription, account: boolean)}
  <li>
    <span>{s.name}</span>
    <select
      class="input"
      value={s.visibility}
      disabled={readOnly}
      aria-label={s.name}
      onchange={(e) => share(s, e.currentTarget.value, account)}
    >
      <option value="private">{tr('only_me')}</option>
      {#each groups as g (g.id)}<option value="group:{g.id}">{tr('group', { name: g.name })}</option>{/each}
      <option value="everyone">{tr('everyone')}</option>
    </select>
  </li>
{/snippet}

<h3>{tr('subscriptions')}</h3>
<ul>
  {#each list as s (s.id)}
    {@render row(s, false)}
  {:else}
    <li class="muted">{tr('none')}</li>
  {/each}
</ul>
{#if accounts.length}
  <h3>{tr('accounts')}</h3>
  <p class="muted small">{tr('accounts_hint')}</p>
  <ul>
    {#each accounts as a (a.id)}
      {@render row(a, true)}
    {/each}
  </ul>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  ul {
    list-style: none;
    padding: 0;
  }
  li {
    display: flex;
    gap: var(--space-2);
    justify-content: space-between;
    align-items: center;
    padding: var(--space-1) 0;
  }
  select {
    max-width: 16rem;
  }
  h3 {
    font-size: 1rem;
    margin: var(--space-3) 0 var(--space-1);
  }
  .small {
    font-size: 0.85em;
  }
</style>
