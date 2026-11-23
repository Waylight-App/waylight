<script lang="ts">
  import { intlLocale, serverMessage, t } from '../../../shared/i18n';
  import { type AttentionItem, groupItems } from '../domain/attention';
  import type { AttentionState } from './attentionState.svelte';

  /** Props of the needs-your-attention page (#592). */
  interface Props {
    state: AttentionState;
    /** Display name of an integration, e.g. "Immich" for "immich". */
    nameOf: (integration: string) => string;
    /** Name of a tab, to say where a widget is; undefined when unknown. */
    tabName?: (tabId: string) => string | undefined;
    /** Runs the entry's action; "check" is handled here. */
    onaction: (item: AttentionItem) => void;
  }
  let { state, nameOf, tabName = () => undefined, onaction }: Props = $props();

  const groups = $derived(groupItems(state.items));
  const since = (d: Date) =>
    new Intl.DateTimeFormat(intlLocale(), { dateStyle: 'medium', timeStyle: 'short' }).format(d);

  // The list is fresh when the page opens and stays fresh while it is open,
  // so a fix made in another window shows up without a reload.
  $effect(() => {
    void state.refresh();
    const id = setInterval(() => void state.refresh(), 15_000);
    return () => clearInterval(id);
  });

  /** Where the entry is and since when, e.g. "On tab Home · since 21 Nov 2026, 10:00". */
  function meta(item: AttentionItem, where: string | undefined): string {
    const place = where
      ? t('attention.on_tab', { value: where })
      : item.widgetId
        ? ''
        : t('attention.instance');
    return [place, t('attention.since', { value: since(item.since) })].filter(Boolean).join(' · ');
  }

  function run(item: AttentionItem) {
    if (item.action === 'check') void state.check();
    else onaction(item);
  }
</script>

<p class="muted">{t('attention.intro')}</p>
{#if state.error}<p class="error" role="alert">{state.error}</p>{/if}

{#if !state.loaded}
  <p class="muted" role="status">{t('attention.loading')}</p>
{:else if state.count === 0}
  <p class="empty" role="status">{t('attention.empty')}</p>
{:else}
  {#each [{ id: 'failing', label: t('attention.failing'), items: groups.failing }, { id: 'setup', label: t('attention.setup'), items: groups.setup }] as g (g.id)}
    {#if g.items.length}
      <section aria-labelledby="attention-{g.id}">
        <h3 id="attention-{g.id}">{g.label} <span class="count">{g.items.length}</span></h3>
        <ul>
          {#each g.items as item (item.id)}
            {@const where = item.tabId ? tabName(item.tabId) : undefined}
            <li data-kind={item.kind}>
              <span class="dot" aria-hidden="true"></span>
              <div class="text">
                <strong>{nameOf(item.integration)}</strong>
                <span>{serverMessage(item.reason.messageKey, item.reason.params) ?? item.reason.message}</span
                >
                <span class="meta muted">{meta(item, where)}</span>
              </div>
              <button
                class="btn"
                class:btn-primary={item.kind === 'connection'}
                type="button"
                disabled={item.action === 'check' && state.checking}
                aria-label="{t(`attention.action.${item.action}`)}: {nameOf(item.integration)}"
                onclick={() => run(item)}>{t(`attention.action.${item.action}`)}</button
              >
            </li>
          {/each}
        </ul>
      </section>
    {/if}
  {/each}
  {#if groups.failing.length}
    <button class="btn" type="button" disabled={state.checking} onclick={() => void state.check()}
      >{state.checking ? t('attention.loading') : t('attention.check_all')}</button
    >
  {/if}
{/if}

<style>
  section {
    margin-bottom: var(--space-4);
  }
  h3 {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .count {
    font-size: var(--fs-xs);
    font-weight: 600;
    padding: 0 var(--space-2);
    border-radius: 999px;
    background: var(--color-surface-2);
  }
  ul {
    list-style: none;
    margin: 0;
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
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    background: var(--color-surface);
  }
  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--color-warning);
  }
  li[data-kind='connection'] .dot {
    background: var(--color-danger);
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
    gap: var(--space-1);
  }
  .meta {
    font-size: var(--fs-xs);
  }
  .empty {
    padding: var(--space-4);
    text-align: center;
    border: 1px dashed var(--color-border);
    border-radius: var(--radius-sm);
  }
  /* Phones: the button moves under the text. */
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
