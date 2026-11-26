<script lang="ts">
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import {
    canAct,
    type ContainerList,
    type DockerActions,
    filterContainers,
    indicator,
    parseDockerConfig,
  } from '../domain/container';
  import { type ImageUpdate, updatesByContainer } from '../domain/imageUpdate';
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import StatusDot, { type StatusLevel } from '../../../shared/ui/StatusDot.svelte';
  import type { DockerUseCases } from '../usecase/dockerUseCases';

  /** Props of the docker widget. */
  interface Props {
    docker: DockerUseCases;
    config: Record<string, unknown>;
  }
  let { docker, config }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 20_000;
  const cfg = $derived(parseDockerConfig(config));
  let data = $state<ContainerList | null>(null);
  let error = $state('');
  const shown = $derived(data ? filterContainers(data.containers, cfg) : []);
  let actions = $state<DockerActions | null>(null);
  let busy = $state('');
  let logs = $state<{ name: string; lines: string[] | null; error: string } | null>(null);
  let updates = $state(new Map<string, ImageUpdate>());

  // Image update hints (#107); silently absent while disabled.
  $effect(() => {
    const load = () =>
      docker
        .updates()
        .then((l) => (updates = updatesByContainer(l ?? [])))
        .catch(() => {});
    return poll(load, 10 * 60_000);
  });

  $effect(() => {
    docker
      .actions()
      .then((a) => (actions = a))
      .catch(() => {});
  });

  async function restart(name: string) {
    if (!confirm(t('docker.confirm_restart', { name }))) return;
    busy = name;
    try {
      await docker.restart(name);
      data = await docker.containers();
      error = '';
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = '';
    }
  }

  async function showLogs(name: string) {
    logs = { name, lines: null, error: '' };
    try {
      const lines = await docker.logs(name);
      if (logs?.name === name) logs = { name, lines, error: '' };
    } catch (err) {
      if (logs?.name === name) logs = { name, lines: [], error: errorMessage(err) };
    }
  }

  const dotLevel: Record<string, StatusLevel> = {
    ok: 'ok',
    warning: 'warning',
    error: 'failed',
    off: 'stopped',
  };
  // Docker writes "Up 3 days"; the row value reads as a quiet phrase like the other rows.
  const phrase = (s: string) => s.charAt(0).toLowerCase() + s.slice(1);

  $effect(() => {
    const load = () =>
      docker
        .containers()
        .then((d) => ((data = d), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });
</script>

{#if error}<p class="error">{error}</p>{/if}
{#if data}
  <p class="summary" class:alert={data.unhealthy > 0}>
    {data.running}
    {t('docker.running')} · {data.stopped}
    {t('docker.stopped')}{#if data.unhealthy}
      · {data.unhealthy} {t('docker.unhealthy')}{/if}
  </p>
  {#if data.error}<p class="error small">{data.error}</p>{/if}
  <ul class="list">
    {#each shown as c (c.id)}
      {@const u = updates.get(c.name)}
      <ListRow nowrap title="{c.image}\n{c.status}">
        {#snippet lead()}<StatusDot level={dotLevel[indicator(c)] ?? 'stopped'} label={c.state} />{/snippet}
        {c.name}
        {#snippet value()}<span class="status">{phrase(c.status)}</span>{/snippet}
        {#snippet trail()}
          {#if u}
            <a
              class="update"
              href={u.releaseUrl || undefined}
              target="_blank"
              rel="noopener noreferrer"
              title={u.newerTag ? t('docker.newer_tag', { tag: u.newerTag }) : t('docker.update_available')}
              aria-label={t('docker.update_available_for', { name: c.name })}
              ><Icon name="upload" size={16} /></a
            >
          {/if}
          {#if canAct(actions, c.name)}
            <button
              data-action
              class="btn btn-icon act"
              type="button"
              title={t('docker.show_logs')}
              aria-label={t('docker.logs_of', { name: c.name })}
              onclick={() => showLogs(c.name)}>≡</button
            >
            <button
              data-action
              class="btn btn-icon act"
              type="button"
              title={t('docker.restart')}
              aria-label={t('docker.restart_2', { name: c.name })}
              disabled={busy === c.name}
              onclick={() => restart(c.name)}>↻</button
            >
          {/if}
        {/snippet}
      </ListRow>
    {:else}
      <li class="muted empty">{cfg.problemsOnly ? t('docker.all_healthy') : t('docker.no_containers')}</li>
    {/each}
  </ul>
{/if}

{#if logs}
  <Modal title={t('docker.logs', { name: logs.name })} onclose={() => (logs = null)}>
    {#if logs.error}<p class="error">{logs.error}</p>{/if}
    {#if logs.lines === null}
      <p class="muted">{t('docker.loading')}</p>
    {:else}
      <pre class="logs">{logs.lines.join('\n') || t('docker.no_output')}</pre>
    {/if}
  </Modal>
{/if}

<style>
  .update {
    color: var(--color-warning);
    text-decoration: none;
    font-size: var(--fs-xs);
  }
  .act {
    padding: 0 var(--space-1);
    line-height: 1.2;
    font-size: var(--fs-sm);
  }
  .logs {
    max-height: 60vh;
    max-width: min(90vw, 900px);
    overflow: auto;
    font-size: var(--fs-xs);
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--color-surface-2);
    padding: var(--space-2);
    border-radius: var(--radius-sm);
  }
  .summary {
    margin: 0 0 var(--space-2);
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    color: var(--color-text-subtle);
  }
  .summary.alert {
    color: var(--color-danger);
  }
  .small {
    font-size: var(--fs-xs);
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .empty {
    padding: var(--space-1) 0;
  }
  .status {
    white-space: nowrap;
  }
</style>
