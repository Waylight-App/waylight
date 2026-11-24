<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { formatBytes } from '../../../shared/ui/format';
  import { handshakeAgo, type PeerList } from '../domain/peer';
  import type { WireGuardUseCases } from '../usecase/wireguardUseCases';

  /** Props of the WireGuard widget. */
  interface Props {
    wireguard: WireGuardUseCases;
  }
  let { wireguard }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 60_000;
  let list = $state<PeerList | null>(null);
  let error = $state('');
  let missing = $state(false);
  let now = $state(new Date());

  $effect(() => {
    const load = () =>
      wireguard
        .peers()
        .then((l) => ((list = l), (now = new Date()), (error = ''), (missing = false)))
        .catch((err) => {
          missing = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });
</script>

{#if error && !list}
  <p class={missing ? 'muted' : 'error'}>{error}</p>
{:else if list}
  <p class="summary">
    <strong>{list.online}</strong>
    {t('wireguard.of_connected', { length: list.peers.length })}
  </p>
  <ul class="peers">
    {#each list.peers as p (p.publicKey)}
      <ListRow nowrap title="{p.allowedIps}{p.endpoint ? ` · ${p.endpoint}` : ''}">
        {#snippet lead()}
          <span
            class="dot"
            class:on={p.online}
            aria-label={p.online ? t('wireguard.online') : t('wireguard.offline')}
          ></span>
        {/snippet}
        {p.name}
        {#snippet value()}
          <span class="small">{handshakeAgo(p.latestHandshake, now)}</span>
          <span class="small traffic">↓{formatBytes(p.receivedBytes)} ↑{formatBytes(p.sentBytes)}</span>
        {/snippet}
      </ListRow>
    {:else}
      <li class="muted">{t('wireguard.no_peers')}</li>
    {/each}
  </ul>
  {#if list.stale}<p class="error small">{t('wireguard.the_export_is_outdated_is_the')}</p>{/if}
{:else}
  <p class="muted">{t('wireguard.loading')}</p>
{/if}

<style>
  .summary {
    margin: 0 0 var(--space-2);
    font-size: var(--fs-sm);
  }
  .peers {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .small {
    font-size: var(--fs-xs);
    white-space: nowrap;
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex: none;
    background: var(--color-border);
  }
  .dot.on {
    background: var(--color-ok);
  }
  @container widget (max-width: 260px) {
    .traffic {
      display: none;
    }
  }
</style>
