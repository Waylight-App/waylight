<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import {
    type AttentionEntry,
    type AttentionSource,
    attentionEntries,
    byUrgency,
    sourceEntry,
    type SourceState,
  } from '../domain/attention';
  import type { AttentionUseCases } from '../usecase/attentionUseCases';

  /**
   * Props of the attention line (#593): the most urgent problem of the tab
   * in one line at the top, with a count of the others. It disappears when
   * all is well.
   */
  interface Props {
    attention: AttentionUseCases;
    /** Reads the short state of a status widget's module (#597). */
    sourceState: (module: string) => Promise<SourceState>;
    tabId: string;
    sources: AttentionSource[];
    /** Shows the widget an entry belongs to. */
    onview: (widgetId: string) => void;
  }
  let { attention, sourceState, tabId, sources, onview }: Props = $props();

  const REFRESH_MS = 60_000;
  let entries = $state<AttentionEntry[]>([]);

  /** Reads every source; one that fails is left out instead of hiding the rest. */
  async function load() {
    const wanted = sources;
    const [items, ...states] = await Promise.allSettled([
      attention.list(),
      ...wanted.map((s) => sourceState(s.module)),
    ]);
    const next: AttentionEntry[] = [];
    states.forEach((r, i) => {
      const source = wanted[i];
      const entry = source && r.status === 'fulfilled' ? sourceEntry(source, r.value) : null;
      if (entry) next.push(entry);
    });
    if (items.status === 'fulfilled') next.push(...attentionEntries(items.value, tabId));
    entries = byUrgency(next);
  }

  const poll = usePolling();
  $effect(() => poll(load, REFRESH_MS));

  const first = $derived(entries[0]);
</script>

{#if first}
  <div class="attention" data-level={first.level} role="status" data-testid="attention-line">
    <span class="dot" aria-hidden="true"></span>
    <span class="text">
      <strong>{first.title}:</strong>
      {first.text}
      {#if entries.length > 1}<span class="more">{t('attention.more', { count: entries.length - 1 })}</span
        >{/if}
    </span>
    {#if first.widgetId}
      {@const id = first.widgetId}
      <button class="btn" type="button" onclick={() => onview(id)}
        >{t('attention.view')}<span class="sr-only">: {first.title}</span></button
      >
    {/if}
  </div>
{/if}

<style>
  .attention {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius);
    border: 1px solid var(--color-border);
    background: var(--color-surface);
  }
  /* Amber or red only for what needs action (#593). */
  .attention[data-level='attention'] {
    border-color: var(--color-warning);
  }
  .attention[data-level='problem'] {
    border-color: var(--color-danger);
  }
  .dot {
    flex: none;
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
  }
  [data-level='attention'] .dot {
    background: var(--color-warning);
  }
  [data-level='problem'] .dot {
    background: var(--color-danger);
  }
  .text {
    flex: 1;
    min-width: 0;
  }
  .more {
    margin-left: var(--space-1);
    color: var(--color-text-muted);
  }
</style>
