<script lang="ts">
  import { t } from '../shared/i18n';
  import { undoQueue } from '../shared/ui/undo.svelte';
  // The toast of the quick actions in widgets (#867): what happened, and Undo.
</script>

{#if undoQueue.current}
  {#key undoQueue.current.key}
    <div class="toast" data-testid="undo-toast">
      <span>{undoQueue.current.text}</span>
      <button class="btn" type="button" aria-keyshortcuts="Control+Z" onclick={() => undoQueue.undo()}
        >{t('dashboard.undo')}</button
      >
    </div>
  {/key}
{/if}

<style>
  /* The same place and look as the undo toast of edit mode. */
  .toast {
    position: fixed;
    left: 50%;
    bottom: calc(var(--space-4) + var(--bottom-nav-h, 0px));
    transform: translateX(-50%);
    display: flex;
    gap: var(--space-3);
    align-items: center;
    max-width: calc(100vw - 2 * var(--space-4));
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-sm);
    background: var(--color-surface);
    color: var(--color-text);
    border: 1px solid var(--color-border);
    box-shadow: 0 4px 16px rgb(0 0 0 / 20%);
    z-index: 20;
  }
  span {
    min-width: 0;
    overflow-wrap: anywhere;
  }
</style>
