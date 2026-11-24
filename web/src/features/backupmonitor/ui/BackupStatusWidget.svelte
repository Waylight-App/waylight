<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import StatusDot from '../../../shared/ui/StatusDot.svelte';
  import { ago, type BackupSource } from '../domain/source';
  import type { BackupMonitorUseCases } from '../usecase/backupMonitorUseCases';

  /** Props of the backup status widget. */
  interface Props {
    backupMonitor: BackupMonitorUseCases;
  }
  let { backupMonitor }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 60_000;
  let sources = $state<BackupSource[]>([]);
  let error = $state('');
  let now = $state(new Date());

  $effect(() => {
    const load = () =>
      backupMonitor
        .list()
        .then((s) => ((sources = s), (now = new Date()), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });

  const labels = {
    ok: 'backupmonitor.state.ok',
    overdue: 'backupmonitor.state.overdue',
    failed: 'backupmonitor.state.failed',
    waiting: 'backupmonitor.state.waiting',
    running: 'backupmonitor.state.running',
  } as const;
  const backupLevel = {
    ok: 'ok',
    running: 'ok',
    waiting: 'stopped',
    overdue: 'warning',
    failed: 'failed',
  } as const;
</script>

{#if error}<p class="error">{error}</p>{/if}
<ul class="list">
  {#each sources as s (s.id)}
    <ListRow nowrap title={s.lastMessage}>
      {#snippet lead()}
        <StatusDot
          level={backupLevel[s.state]}
          label={t(labels[s.state])}
          word={s.state !== 'ok' && s.state !== 'running'}
        />
      {/snippet}
      {s.name}
      {#snippet value()}
        <span class="when"
          >{s.state === 'failed'
            ? t('backupmonitor.failed_ago', { time: ago(s.lastFailureAt, now) })
            : ago(s.lastSuccessAt, now)}</span
        >
      {/snippet}
    </ListRow>
  {:else}
    <li class="muted">{t('backupmonitor.no_backups_monitored_yet_add_them')}</li>
  {/each}
</ul>

<style>
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .when {
    font-size: var(--fs-xs);
    white-space: nowrap;
  }
</style>
