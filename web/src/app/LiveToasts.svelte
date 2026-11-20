<script lang="ts">
  import { t } from '../shared/i18n';
  import { dismissLiveAlert, liveAlerts } from './liveAlerts.svelte';

  /** Live notifications of the browser channel as toasts (#583). */
</script>

<div class="toasts" aria-live="polite">
  {#each liveAlerts.list as a (a.key)}
    <div
      class="toast"
      class:critical={a.severity === 'critical'}
      role={a.severity === 'critical' ? 'alert' : 'status'}
    >
      <span class="text"
        ><strong>{a.title}</strong>{#if a.body}<span class="body">{a.body}</span>{/if}</span
      >
      <button class="btn" type="button" onclick={() => dismissLiveAlert(a.key)}
        >{t('notifications.live_dismiss')}</button
      >
    </div>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    top: var(--space-4);
    right: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    z-index: 60;
    width: min(24rem, calc(100vw - 2 * var(--space-4)));
  }
  .toast {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    background: var(--color-surface);
    color: var(--color-text);
    border: 1px solid var(--color-border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }
  .toast.critical {
    border-color: var(--color-danger);
  }
  .text {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .body {
    font-size: 0.9rem;
    color: var(--color-text-muted);
  }
</style>
