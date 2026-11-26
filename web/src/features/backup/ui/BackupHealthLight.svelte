<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { BackupStatus } from '../domain/backup';
  import StatusDot from '../../../shared/ui/StatusDot.svelte';
  import { healthText } from './healthText';

  /** Props of the backup traffic light (#677). */
  interface Props {
    status: BackupStatus;
    now?: Date;
    /** Also shows the server's reason for a failure. */
    details?: boolean;
  }
  let { status, now = new Date(), details = false }: Props = $props();

  const line = $derived(healthText(status, now));
  // The sentence names the state, so the dot needs no word of its own.
  const level = $derived(line.tone === 'error' ? 'failed' : line.tone === 'warn' ? 'warning' : 'ok');
</script>

<div class="light" role="status">
  <span class="dot"><StatusDot {level} /></span>
  <span class="text">
    <span>{line.text}</span>
    {#if details && status.health === 'failed' && status.error}
      <span class="reason">{t('backup.health.error', { error: status.error })}</span>
    {/if}
    {#if details && status.offsiteError}
      <span class="reason">{t('backup.health.offsite_error', { error: status.offsiteError })}</span>
    {/if}
  </span>
</div>

<style>
  .light {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
  }
  /* Centres the 8 px dot on the first line of text. */
  .dot {
    display: flex;
    height: var(--lh-sm);
  }
  .text {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }
  .reason {
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
    overflow-wrap: anywhere;
  }
</style>
