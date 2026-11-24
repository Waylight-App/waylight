<script lang="ts">
  import { t } from '../../../shared/i18n';
  import StatusDot from '../../../shared/ui/StatusDot.svelte';
  import type { AttentionEntry } from '../domain/attentionLine';

  /**
   * Props of the alert bar (#593, #770): the most urgent entry of the tab at
   * the top with a link to its widget, and a count of the others. Nothing is rendered
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
  <div class="alert" data-level={first.level} role="status" data-testid="attention-line">
    <StatusDot
      level={first.level === 'problem' ? 'failed' : 'warning'}
      label={t(first.level === 'problem' ? 'compact.level_problem' : 'compact.level_attention')}
      word
    />
    <span class="text">
      <strong>{first.title}:</strong>
      {first.text}
      {#if entries.length > 1}<span class="more"
          >{t('dashboard.attention_more', { count: entries.length - 1 })}</span
        >{/if}
    </span>
    {#if first.widgetId}
      {@const id = first.widgetId}
      <button class="link" type="button" onclick={() => onview(id)}
        >{t('dashboard.attention_view')}<span class="sr-only">: {first.title}</span></button
      >
    {/if}
  </div>
{/if}

<style>
  /* One bar for the most important failure; a quiet tint of its status, no glow. */
  .alert {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    border: 1px solid var(--color-border);
    background: var(--color-surface);
  }
  .alert[data-level='attention'] {
    border-color: var(--color-warning);
    background: color-mix(in srgb, var(--color-warning) 8%, var(--color-surface));
  }
  .alert[data-level='problem'] {
    border-color: var(--color-danger);
    background: color-mix(in srgb, var(--color-danger) 8%, var(--color-surface));
  }
  .alert :global(.word) {
    color: var(--color-text);
    font-weight: 600;
  }
  .text {
    flex: 1;
    min-width: 0;
  }
  .more {
    margin-left: var(--space-1);
    color: var(--color-text-muted);
  }
  .link {
    flex: none;
    min-height: 44px;
    padding: 0 var(--space-2);
    border: 0;
    background: none;
    color: var(--color-accent-text);
    font: inherit;
    font-weight: 600;
    text-decoration: underline;
    text-underline-offset: 3px;
    cursor: pointer;
  }
  @media (max-width: 640px) {
    .alert {
      flex-wrap: wrap;
      gap: var(--space-1) var(--space-3);
    }
    .text {
      flex-basis: 100%;
      order: 2;
    }
    .link {
      order: 1;
      margin-left: auto;
    }
  }
</style>
