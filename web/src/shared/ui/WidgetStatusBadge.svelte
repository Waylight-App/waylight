<script lang="ts">
  import Icon from './icons/Icon.svelte';
  import { type Freshness, minutesAgo } from '../freshness/status';
  import { intlLocale, t } from '../i18n';

  /** Props of the shared data freshness badge (#94). */
  interface Props {
    freshness: Freshness;
    fetchedAt: number | null;
    lastError: string | null;
    now: number;
    onrefresh: () => void;
    /** In a card's header row (#851); otherwise it floats over the bottom right of the content. */
    inline?: boolean;
  }
  let { freshness, fetchedAt, lastError, now, onrefresh, inline = false }: Props = $props();

  const age = $derived.by(() => {
    if (fetchedAt === null) return t('freshness.not_loaded');
    const m = minutesAgo(fetchedAt, now);
    return m === 0 ? t('freshness.just_now') : t('freshness.minutes_ago', { minutes: m });
  });
  const title = $derived(
    [
      fetchedAt === null
        ? ''
        : t('freshness.updated_at', { time: new Date(fetchedAt).toLocaleString(intlLocale()) }),
      lastError ? t('freshness.last_error', { error: lastError }) : '',
    ]
      .filter(Boolean)
      .join('\n'),
  );
</script>

<div class="badge {freshness}" class:inline title={title || undefined}>
  {#if freshness === 'error'}<Icon name="warning" size={16} />{/if}
  <span class="age">
    {freshness === 'stale'
      ? t('freshness.stale', { age })
      : freshness === 'error'
        ? t('freshness.failed', { age })
        : age}
  </span>
  <button
    class="refresh"
    type="button"
    onclick={onrefresh}
    aria-label={t('freshness.refresh')}
    title={t('freshness.refresh')}>↻</button
  >
</div>

<style>
  .badge {
    position: absolute;
    right: 4px;
    bottom: 2px;
    display: flex;
    gap: var(--space-1);
    align-items: center;
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
    /* A chip of its own surface, so the text keeps 4.5:1 over any card,
       including the lit one whose colours invert the scheme. */
    padding: 0 var(--space-2);
    border-radius: 999px;
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    /* visibility (not only opacity) keeps the hidden badge away from screen
       readers; it switches at the end of the fade-out. */
    opacity: 0;
    visibility: hidden;
    transition:
      opacity 0.15s,
      visibility 0.15s;
    pointer-events: none;
  }
  /* Shown on hover/focus; stale data and errors are always visible. */
  :global(.lazy-root:hover) .badge,
  :global(.lazy-root:focus-within) .badge,
  .badge.stale,
  .badge.error {
    opacity: 1;
    visibility: visible;
    pointer-events: auto;
  }
  /* In the header row the badge takes its own place after the meta text,
     which shrinks instead of being covered. display (not visibility) frees
     the room while it is hidden, so the meta keeps its full width then. */
  .badge.inline {
    position: static;
    flex: none;
    align-self: center;
    margin: -4px 0;
    display: none;
    opacity: 1;
    visibility: visible;
    pointer-events: auto;
  }
  :global(.widget:hover) .badge.inline,
  :global(.widget:focus-within) .badge.inline,
  .badge.inline.stale,
  .badge.inline.error {
    display: flex;
  }
  .badge.stale {
    color: var(--color-warning);
  }
  .badge.error {
    color: var(--color-danger);
  }
  /* 24 × 24 px hit area around the small glyph (WCAG 2.5.8, #317). */
  .refresh {
    display: inline-grid;
    place-items: center;
    min-width: 24px;
    min-height: 24px;
    border: 0;
    background: none;
    color: inherit;
    cursor: pointer;
    padding: 0;
    font: inherit;
    font-size: var(--fs-sm);
  }
  @media (prefers-reduced-motion: reduce) {
    .badge {
      transition: none;
    }
  }
</style>
