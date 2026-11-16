<script lang="ts">
  import type { FeatureLevel, PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { features } from './features';
  import { LicenceApi, statusKey, type Licence } from './licence';
  import { texts } from './texts';

  /**
   * Settings → Features for organisations (#307, ADR 0021): every feature
   * with its state, and the licence key for admins.
   */
  interface Props {
    context: PluginContext;
  }
  let { context }: Props = $props();

  const api = $derived(new LicenceApi(context.http));
  const tr = (key: string, params: Record<string, string | number> = {}) =>
    text(texts, key, context.language, params);
  const label = (t: { en: string } & Record<string, string>) => t[context.language] ?? t.en;

  let levels = $state<Record<string, FeatureLevel>>({});
  let licence = $state<Licence | null>(null);
  let key = $state('');
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  async function load() {
    try {
      levels = await api.levels();
      if (context.account.admin) licence = await api.get();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  $effect(() => {
    load();
  });

  async function install(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      licence = await api.install(key);
      key = '';
      message = tr('saved');
      levels = await api.levels();
    } catch (err) {
      error = context.errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove() {
    if (!confirm(tr('remove_confirm'))) return;
    error = message = '';
    try {
      await api.remove();
      await load();
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  const stateLabel = (l: FeatureLevel | undefined) =>
    l === 'full' ? tr('on') : l === 'read' ? tr('read') : tr('off');
</script>

<p class="muted">{tr('intro')}</p>

<ul class="features">
  {#each features as f (f.id)}
    {@const level = levels[f.id]}
    <li class:off={!level || level === 'none'}>
      <div>
        <strong>{label(f.name)}</strong>
        <span class="muted small">{label(f.description)}</span>
      </div>
      <span class="state state-{level ?? 'none'}">{stateLabel(level)}</span>
    </li>
  {/each}
</ul>

<h3>{tr('key_title')}</h3>
{#if !context.account.admin}
  <p class="muted">{tr('admins_only')}</p>
{:else if licence}
  <p role="status">
    {tr(statusKey(licence), {
      licensee: licence.licensee ?? '',
      until: licence.until ?? '',
      seats: licence.seats,
      accounts: licence.accounts,
      problem: licence.problem ?? '',
    })}
  </p>
  {#if licence.status !== 'missing' && licence.seats > 0}
    <p class="muted small">{tr('seats', { accounts: licence.accounts, seats: licence.seats })}</p>
  {/if}
  {#if licence.unknown.length}<p class="muted small">
      {tr('unknown', { list: licence.unknown.join(', ') })}
    </p>{/if}
  {#if licence.source === 'environment'}
    <p class="muted">{tr('from_env')}</p>
  {:else}
    <form onsubmit={install}>
      <textarea
        class="input"
        rows="4"
        bind:value={key}
        placeholder={tr('paste')}
        aria-label={tr('paste')}
        required></textarea>
      <div class="actions">
        <button class="btn btn-primary" type="submit" disabled={busy}>{tr('install')}</button>
        {#if licence.source === 'settings'}
          <button class="btn btn-danger" type="button" onclick={remove}>{tr('remove')}</button>
        {/if}
      </div>
    </form>
  {/if}
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  .features {
    list-style: none;
    padding: 0;
    margin: 0 0 var(--space-3);
  }
  .features li {
    display: flex;
    gap: var(--space-2);
    justify-content: space-between;
    align-items: center;
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border, rgba(127, 127, 127, 0.2));
  }
  .features li div {
    display: flex;
    flex-direction: column;
  }
  .features li.off strong {
    opacity: 0.7;
  }
  .small {
    font-size: 0.85em;
  }
  .state {
    font-size: 0.85em;
    padding: 0 var(--space-2);
    border-radius: 999px;
    white-space: nowrap;
  }
  .state-full {
    background: var(--color-success-bg, rgba(34, 197, 94, 0.15));
  }
  .state-read {
    background: var(--color-warning-bg, rgba(234, 179, 8, 0.18));
  }
  .state-none {
    background: var(--color-surface-2, rgba(127, 127, 127, 0.12));
  }
  form {
    display: grid;
    gap: var(--space-2);
  }
  textarea {
    font-family: var(--font-mono, monospace);
    font-size: 0.85em;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
  h3 {
    font-size: 1rem;
    margin: var(--space-3) 0 var(--space-2);
  }
</style>
