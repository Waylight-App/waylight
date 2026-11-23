<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { AttentionState } from './attentionState.svelte';

  /** Props of the top bar dot (#592). */
  interface Props {
    state: AttentionState;
    onopen: () => void;
  }
  let { state, onopen }: Props = $props();

  const label = $derived(
    state.count === 1 ? t('attention.badge_one') : t('attention.badge', { n: state.count }),
  );
</script>

<!-- Hidden while everything works, so it only draws the eye when it matters. -->
{#if state.count > 0}
  <button type="button" class="badge" aria-label={label} title={label} onclick={onopen}>
    <span class="dot" aria-hidden="true"></span>{state.count}
  </button>
{/if}

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: 999px;
    background: var(--color-surface);
    color: var(--color-text);
    font: inherit;
    font-weight: 600;
    cursor: pointer;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-warning);
  }
</style>
