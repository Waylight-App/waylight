<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import StatusDot from '../../../shared/ui/StatusDot.svelte';
  import { useHeaderMeta } from '../../../shared/ui/widgetHeaderMeta';
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

  // "2 of 3 connected" is the card's header meta (#809).
  const setMeta = useHeaderMeta();
  $effect(() =>
    setMeta(
      list ? { text: `${list.online} ${t('wireguard.of_connected', { length: list.peers.length })}` } : null,
    ),
  );
</script>

{#if error && !list}
  <p class={missing ? 'muted' : 'error'}>{error}</p>
{:else if list}
  <ul class="peers">
    {#each list.peers as p (p.publicKey)}
      <ListRow nowrap title="{p.allowedIps}{p.endpoint ? ` · ${p.endpoint}` : ''}">
        {#snippet lead()}<StatusDot
            level={p.online ? 'ok' : 'stopped'}
            label={p.online ? t('wireguard.online') : t('wireguard.offline')}
          />{/snippet}
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
  .peers {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .small {
    font-size: var(--fs-xs);
    white-space: nowrap;
  }
  @container widget (max-width: 260px) {
    .traffic {
      display: none;
    }
  }
</style>
