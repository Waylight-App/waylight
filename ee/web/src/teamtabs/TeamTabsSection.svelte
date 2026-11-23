<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { TeamTabsApi, tabsFor, type Group, type TeamTab } from './teamtabs';
  import { texts } from './texts';

  /** Settings → Team tabs (#295): share own tabs with a group or everyone. */
  interface Props {
    context: PluginContext;
  }
  let { context }: Props = $props();

  const api = $derived(new TeamTabsApi(context.http));
  const tr = (key: string, params: Record<string, string | number> = {}) =>
    text(texts, key, context.language, params);
  const readOnly = $derived(context.level('team-tabs') === 'read');

  let tabs = $state<TeamTab[]>([]);
  let groups = $state<Group[]>([]);
  let defaults = $state<Record<string, string>>({});
  let error = $state('');
  let message = $state('');

  $effect(() => {
    Promise.all([api.tabs(), api.groups(), api.defaults()])
      .then(([t, g, d]) => ((tabs = t), (groups = g), (defaults = d)))
      .catch((err) => (error = context.errorMessage(err)));
  });

  async function setDefault(groupId: string, tabId: string) {
    error = message = '';
    try {
      await api.setDefault(groupId, tabId);
      defaults = { ...defaults, [groupId]: tabId };
      message = tr('saved');
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function share(tab: TeamTab, visibility: string) {
    error = message = '';
    try {
      tab.visibility = (await api.share(tab.id, visibility)).visibility;
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
  {#each tabs as tab (tab.id)}
    <li>
      <span>{tab.name}</span>
      <select
        class="input"
        value={tab.visibility}
        disabled={readOnly}
        aria-label={tab.name}
        onchange={(e) => share(tab, e.currentTarget.value)}
      >
        <option value="private">{tr('only_me')}</option>
        {#each groups as g (g.id)}<option value="group:{g.id}">{tr('group', { name: g.name })}</option>{/each}
        <option value="everyone">{tr('everyone')}</option>
      </select>
    </li>
  {:else}
    <li class="muted">{tr('no_tabs')}</li>
  {/each}
</ul>
{#if groups.length}
  <h3>{tr('defaults_title')}</h3>
  <p class="muted small">{tr('defaults_hint')}</p>
  <ul>
    {#each groups as g (g.id)}
      <li>
        <span>{g.name}</span>
        <select
          class="input"
          value={defaults[g.id] ?? ''}
          disabled={readOnly}
          aria-label={tr('defaults_title') + ': ' + g.name}
          onchange={(e) => setDefault(g.id, e.currentTarget.value)}
        >
          <option value="">{tr('no_default')}</option>
          {#each tabsFor(tabs, g.id) as tab (tab.id)}<option value={tab.id}>{tab.name}</option>{/each}
        </select>
      </li>
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
    font-size: var(--fs-sm);
    margin: var(--space-3) 0 var(--space-1);
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
