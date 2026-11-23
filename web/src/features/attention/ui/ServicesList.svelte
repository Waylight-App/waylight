<script lang="ts">
  import { intlLocale, serverMessage, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { connectionCounts, type ServiceConnection, sortConnections } from '../domain/attention';
  import type { AttentionUseCases } from '../usecase/attentionUseCases';

  /** Props of Settings → Services (#591): every connection with its state. */
  interface Props {
    attention: AttentionUseCases;
    nameOf: (integration: string) => string;
    widgetName: (kind: string) => string;
    tabName: (tabId: string) => string | undefined;
    /** The settings section of an integration, if it has one. */
    sectionOf: (integration: string) => string | undefined;
    onopen: (section: string) => void;
  }
  let { attention, nameOf, widgetName, tabName, sectionOf, onopen }: Props = $props();

  let list = $state<ServiceConnection[]>([]);
  let loaded = $state(false);
  let checking = $state(false);
  let error = $state('');

  const sorted = $derived(sortConnections(list, nameOf, intlLocale()));
  const counts = $derived(connectionCounts(list));

  async function load(fresh: boolean) {
    try {
      list = await attention.connections(fresh);
      error = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loaded = true;
    }
  }
  $effect(() => {
    void load(false);
  });

  async function checkAll() {
    checking = true;
    await load(true);
    checking = false;
  }

  function stateText(c: ServiceConnection): string {
    if (c.state === 'ok') return t('services.ok');
    if (c.reason) return serverMessage(c.reason.messageKey, c.reason.params) ?? c.reason.message;
    return t('services.setup');
  }

  function usage(c: ServiceConnection): string {
    if (!c.widgets.length) return t('services.unused');
    const names = c.widgets.map((w) => {
      const tab = tabName(w.tabId);
      return tab ? `${widgetName(w.kind)} (${tab})` : widgetName(w.kind);
    });
    return t('services.used_by', { list: names.join(', ') });
  }

  function label(c: ServiceConnection): string {
    if (c.state === 'ok') return t('services.open');
    return t(`attention.action.${c.action ?? 'connect'}`);
  }
</script>

<p class="muted">{t('services.intro')}</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if !loaded}
  <p class="muted" role="status">{t('attention.loading')}</p>
{:else if list.length === 0}
  <p class="muted">{t('services.empty')}</p>
{:else}
  <p class="summary" role="status">{t('services.summary', counts)}</p>
  <ul>
    {#each sorted as c (c.integration)}
      {@const section = sectionOf(c.integration)}
      <li data-state={c.state}>
        <span class="dot" aria-hidden="true"></span>
        <div class="text">
          <strong>{nameOf(c.integration)}</strong>
          <span>{stateText(c)}</span>
          <span class="muted small">{usage(c)}{c.instance ? ` · ${t('services.instance')}` : ''}</span>
        </div>
        {#if c.action === 'check'}
          <button class="btn" type="button" disabled={checking} onclick={checkAll}>{label(c)}</button>
        {:else if section}
          <button
            class="btn"
            class:btn-primary={c.state === 'failing'}
            type="button"
            aria-label="{label(c)}: {nameOf(c.integration)}"
            onclick={() => onopen(section)}>{label(c)}</button
          >
        {/if}
      </li>
    {/each}
  </ul>
  <button class="btn" type="button" disabled={checking} onclick={checkAll}
    >{checking ? t('attention.loading') : t('services.check_all')}</button
  >
{/if}

<style>
  ul {
    list-style: none;
    margin: 0 0 var(--space-3);
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  li {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
  }
  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--color-border);
  }
  li[data-state='ok'] .dot {
    background: var(--color-ok);
  }
  li[data-state='failing'] .dot {
    background: var(--color-danger);
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .small {
    font-size: var(--fs-xs);
  }
  .summary {
    font-weight: 600;
  }
  @media (max-width: 480px) {
    li {
      grid-template-columns: auto 1fr;
    }
    li .btn {
      grid-column: 2;
      justify-self: start;
    }
  }
</style>
