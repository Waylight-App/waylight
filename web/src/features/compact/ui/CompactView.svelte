<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { announce } from '../../../shared/a11y/live.svelte';
  import {
    COMPACT_REFRESH_MS,
    type CompactChoice,
    compactText,
    type CompactState,
    isAlarming,
  } from '../domain/compact';
  import type { CompactUseCases } from '../usecase/compactUseCases';

  /** Props of a compact widget view (#595): the widget's short state in one or two lines. */
  interface Props {
    compact: CompactUseCases;
    /** Module that serves the state at GET /api/<module>/compact. */
    module: string;
    /** Name of the widget, shown as a small label above the state. */
    title: string;
    /** One line for strip tiles instead of label, state and detail on separate lines. */
    strip?: boolean;
    /** Reports every loaded state, e.g. for the detail panel or the attention line. */
    onstate?: (state: CompactState) => void;
    /** Offers the state's quick action (#598); off for guests and kiosk displays. */
    actions?: boolean;
  }
  let { compact, module, title, strip = false, onstate, actions = false }: Props = $props();

  let current = $state<CompactState | null>(null);
  let error = $state('');
  let busy = $state(false);
  let actionError = $state('');

  async function load() {
    try {
      const next = await compact.state(module);
      current = next;
      error = '';
      onstate?.(next);
    } catch (err) {
      error = errorMessage(err);
      throw err;
    }
  }

  const poll = usePolling();
  $effect(() => poll(load, COMPACT_REFRESH_MS));

  /**
   * Runs the quick action. A chip is shown as selected at once and goes
   * back when the server refuses; the server's message is shown then.
   */
  async function act(choice?: CompactChoice) {
    const state = current;
    const action = state?.action;
    if (!state || !action || busy) return;
    busy = true;
    actionError = '';
    const before = action.choices;
    if (choice) action.choices = before.map((c) => ({ ...c, selected: c.id === choice.id }));
    try {
      await compact.act(action, choice);
      announce(t('compact.action_done', { action: compactText(choice?.label ?? action.label) }));
      await load().catch(() => undefined);
    } catch (err) {
      action.choices = before;
      actionError = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  const level = $derived(current?.level ?? 'none');
</script>

<div class="compact" class:strip data-level={level} data-testid="compact-view">
  <span class="label">{title}</span>
  {#if current}
    <span class="state">
      {#if isAlarming(level)}
        <span class="dot" aria-hidden="true"></span>
        <span class="sr-only">{t(`compact.level_${level as 'attention' | 'problem'}`)}:</span>
      {/if}
      {compactText(current.title)}
    </span>
    {#if current.detail}<span class="detail">{compactText(current.detail)}</span>{/if}
    {#if actions && current.action}
      {@const action = current.action}
      {#if action.choices.length > 0}
        <span class="actions" role="group" aria-label={compactText(action.label)}>
          {#each action.choices as choice (choice.id)}
            <button
              class="chip"
              type="button"
              aria-pressed={choice.selected}
              disabled={busy}
              onclick={() => act(choice)}>{compactText(choice.label)}</button
            >
          {/each}
        </span>
      {:else}
        <span class="actions">
          <button class="btn action" type="button" disabled={busy} onclick={() => act()}
            >{compactText(action.label)}</button
          >
        </span>
      {/if}
    {/if}
    {#if actionError}<span class="action-error" role="alert">{actionError}</span>{/if}
  {:else if error}
    <span class="detail" title={error}>{t('compact.unavailable')}</span>
  {:else}
    <span class="detail">{t('compact.loading')}</span>
  {/if}
</div>

<style>
  .compact {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: var(--space-1);
    min-width: 0;
  }
  .compact.strip {
    flex-direction: row;
    align-items: baseline;
    justify-content: flex-start;
    gap: var(--space-2);
  }
  /* Above the card's stretched "show details" button, so they get the click. */
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
    pointer-events: auto;
    min-width: 0;
  }
  .action {
    max-width: 100%;
    padding: 2px var(--space-2);
    font-size: 0.85em;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chip {
    padding: 2px var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: var(--color-surface);
    color: inherit;
    font: inherit;
    font-size: 0.85em;
    cursor: pointer;
  }
  .chip[aria-pressed='true'] {
    border-color: var(--color-accent);
    background: color-mix(in srgb, var(--color-accent) 15%, transparent);
  }
  .action-error {
    color: var(--color-danger);
    font-size: 0.85em;
    pointer-events: auto;
  }
  .label {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  .label,
  .state,
  .detail {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .state {
    font-size: 1.15em;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  .strip .state {
    font-size: 1em;
  }
  .detail {
    color: var(--color-text-muted);
    font-size: 0.9em;
  }
  /* Colour only for what needs action (#593): amber for attention, red for a problem. */
  .dot {
    flex: none;
    width: 0.6em;
    height: 0.6em;
    border-radius: 50%;
  }
  [data-level='attention'] .dot {
    background: var(--color-warning);
  }
  [data-level='problem'] .dot {
    background: var(--color-danger);
  }
  [data-level='attention'] .detail {
    color: var(--color-warning);
  }
  [data-level='problem'] .detail {
    color: var(--color-danger);
  }
</style>
