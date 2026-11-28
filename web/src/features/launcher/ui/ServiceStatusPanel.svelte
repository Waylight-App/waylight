<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { type LauncherService, safeUrl, statusGroups } from '../domain/launcher';
  import type { LauncherUseCases } from '../usecase/launcherUseCases';

  /**
   * Props of the status popover of the top bar pill (#858): only the state
   * of the services, what is down and what is fine. Opening services and
   * bookmarks is the launcher's job.
   */
  interface Props {
    launcher: LauncherUseCases;
    onclose: () => void;
  }
  let { launcher, onclose }: Props = $props();

  let services = $state<LauncherService[]>([]);
  let loaded = $state(false);
  let panel: HTMLDivElement;

  $effect(() => {
    launcher
      .services()
      .catch(() => [])
      .then((s) => ((services = s), (loaded = true)));
  });

  const groups = $derived(statusGroups(services));
  const sections = $derived(
    [
      { key: 'offline', label: t('launcher.status_down'), items: groups.offline },
      { key: 'unknown', label: t('launcher.status_unchecked'), items: groups.unknown },
      { key: 'online', label: t('launcher.status_fine'), items: groups.online },
    ].filter((s) => s.items.length > 0),
  );

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      onclose();
    }
  }

  // Clicks outside the panel close it (the triggers toggle it themselves).
  $effect(() => {
    const away = (e: PointerEvent) => {
      const target = e.target as Node | null;
      if (target && !panel.contains(target) && !(target as Element).closest?.('[data-launcher-trigger]'))
        onclose();
    };
    document.addEventListener('pointerdown', away);
    return () => document.removeEventListener('pointerdown', away);
  });

  // The dialog takes the focus so Esc works right after opening.
  $effect(() => panel.focus());
</script>

<div
  class="panel card"
  role="dialog"
  tabindex="-1"
  aria-label={t('launcher.status_title')}
  bind:this={panel}
  {onkeydown}
>
  {#if !loaded}
    <p class="muted">{t('common.loading')}</p>
  {:else if sections.length === 0}
    <p class="muted">{t('launcher.none_services')}</p>
  {/if}
  {#each sections as s (s.key)}
    <section aria-labelledby="status-{s.key}">
      <h3 id="status-{s.key}">{s.label} <span class="count">{s.items.length}</span></h3>
      <ul>
        {#each s.items as svc (svc.id)}
          {@const href = safeUrl(svc.url)}
          <li data-state={svc.state}>
            <span class="dot" aria-hidden="true"></span>
            <a href={href ?? undefined} target="_blank" rel="noopener noreferrer">{svc.name}</a>
            <span class="state">
              {#if svc.state === 'online'}{t('launcher.online_ms', {
                  ms: svc.latencyMs,
                })}{:else if svc.state === 'offline'}{t('launcher.offline')}{:else}{t(
                  'launcher.unknown',
                )}{/if}
            </span>
          </li>
        {/each}
      </ul>
    </section>
  {/each}
</div>

<style>
  .panel {
    position: absolute;
    top: calc(100% + var(--space-1));
    right: var(--space-4);
    z-index: 40;
    width: min(24rem, calc(100vw - 2 * var(--space-3)));
    max-height: min(80vh, 36rem);
    overflow: auto;
    padding: var(--space-3);
    display: grid;
    gap: var(--space-3);
    box-shadow: var(--shadow-pop);
    outline: none;
  }
  h3 {
    margin: 0 0 var(--space-2);
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  .count {
    font-weight: 400;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 32px;
  }
  a {
    flex: 1;
    min-width: 0;
    color: inherit;
    text-decoration: none;
    overflow-wrap: anywhere;
  }
  a:hover,
  a:focus-visible {
    text-decoration: underline;
  }
  .state {
    flex: none;
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  [data-state='offline'] .state {
    color: var(--color-danger);
    font-weight: 500;
  }
  .dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-border);
  }
  [data-state='online'] .dot {
    background: var(--color-ok);
  }
  [data-state='offline'] .dot {
    background: var(--color-danger);
  }
  @media (max-width: 600px) {
    .panel {
      right: var(--space-2);
      left: var(--space-2);
      width: auto;
    }
  }
</style>
