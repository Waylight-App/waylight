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
  let groups = $state<Group[]>([]);
  let error = $state('');
  let message = $state('');

  $effect(() => {
    Promise.all([api.subscriptions(), api.groups()])
      .then(([s, g]) => ((list = s), (groups = g)))
      .catch((err) => (error = context.errorMessage(err)));
  });

  async function share(s: Subscription, visibility: string) {
    error = message = '';
    try {
      s.visibility = (await api.share(s.id, visibility)).visibility;
      message = tr('saved');
    } catch (err) {
      error = context.errorMessage(err);
    }
  }
</script>

<p class="muted">{tr('description')}</p>
<p class="muted small">{tr('hint')}</p>
{#if readOnly}<p class="muted">{tr('read_only')}</p>{/if}
<ul>
  {#each list as s (s.id)}
    <li>
      <span>{s.name}</span>
      <select
        class="input"
        value={s.visibility}
        disabled={readOnly}
        aria-label={s.name}
        onchange={(e) => share(s, e.currentTarget.value)}
      >
        <option value="private">{tr('only_me')}</option>
        {#each groups as g (g.id)}<option value="group:{g.id}">{tr('group', { name: g.name })}</option>{/each}
        <option value="everyone">{tr('everyone')}</option>
      </select>
    </li>
  {:else}
    <li class="muted">{tr('none')}</li>
  {/each}
</ul>
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
  .small {
    font-size: 0.85em;
  }
</style>
