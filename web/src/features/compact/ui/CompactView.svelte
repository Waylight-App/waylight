<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { COMPACT_REFRESH_MS, compactText, type CompactState, isAlarming } from '../domain/compact';
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
  }
  let { compact, module, title, strip = false, onstate }: Props = $props();

  let current = $state<CompactState | null>(null);
  let error = $state('');

  const poll = usePolling();
  $effect(() =>
    poll(async () => {
      try {
        const next = await compact.state(module);
        current = next;
        error = '';
        onstate?.(next);
      } catch (err) {
        error = errorMessage(err);
        throw err;
      }
    }, COMPACT_REFRESH_MS),
  );

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
