<script lang="ts">
  import Icon from '../shared/ui/icons/Icon.svelte';
  import { t } from '../shared/i18n';
  import { timerAlerts } from './timerAlerts.svelte';
  import { timerTitle } from './timers';

  /** In-app notice of timers that ran out (#474); works without notification permission. */
  const time = (end: number) => new Date(end).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  const dismiss = (id: string) => (timerAlerts.list = timerAlerts.list.filter((x) => x.id !== id));
</script>

<div class="toasts">
  {#each timerAlerts.list as timer (timer.id)}
    <div class="toast" role="alert">
      <span><Icon name="timer" size={16} /> <strong>{timerTitle(timer)}</strong> · {time(timer.end)}</span>
      <button class="btn" type="button" onclick={() => dismiss(timer.id)}>{t('timers.dismiss')}</button>
    </div>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    bottom: var(--space-4);
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    z-index: 60;
    max-width: calc(100vw - 2 * var(--space-4));
  }
  .toast {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    background: var(--color-surface);
    color: var(--color-text);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow);
  }
</style>
