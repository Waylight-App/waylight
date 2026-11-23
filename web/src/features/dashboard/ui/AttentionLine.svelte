<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { AttentionEntry } from '../domain/attentionLine';

  /**
   * Props of the attention line (#593): the most urgent entry of the tab in
   * one line at the top, with a count of the others. Nothing is rendered
   * when all is well.
   */
  interface Props {
    /** Entries ordered by urgency (see byUrgency). */
    entries: AttentionEntry[];
    /** Shows the widget an entry belongs to. */
    onview: (widgetId: string) => void;
  }
  let { entries, onview }: Props = $props();

  const first = $derived(entries[0]);
</script>

{#if first}
  <div class="attention" data-level={first.level} role="status" data-testid="attention-line">
    <span class="dot" aria-hidden="true"></span>
    <span class="text">
      <strong>{first.title}:</strong>
      {first.text}
      {#if entries.length > 1}<span class="more"
          >{t('dashboard.attention_more', { count: entries.length - 1 })}</span
        >{/if}
    </span>
    {#if first.widgetId}
      {@const id = first.widgetId}
      <button class="btn" type="button" onclick={() => onview(id)}
        >{t('dashboard.attention_view')}<span class="sr-only">: {first.title}</span></button
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
    border-radius: var(--radius-md);
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
