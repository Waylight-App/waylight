<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t, type MessageKey } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { countdown, type Launch, parseSpaceConfig, type Picture, statusTone } from '../domain/space';
  import type { SpaceUseCases } from '../usecase/spaceUseCases';

  /** Props of the space widget (#64). */
  interface Props {
    space: SpaceUseCases;
    config: Record<string, unknown>;
  }
  let { space, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseSpaceConfig(config));

  // The server caches launches for an hour and the picture for three.
  const POLL_MS = 15 * 60_000;
  let launches = $state<Launch[] | null>(null);
  let picture = $state<Picture | null>(null);
  let off = $state(false);
  let error = $state('');
  let now = $state(new Date());

  // 404 = switched off in the settings: that source is simply not shown.
  const optional = <T,>(p: Promise<T>): Promise<T | null> =>
    p.catch((err) => {
      if (err instanceof ApiError && err.status === 404) return null;
      throw err;
    });

  $effect(() => {
    const { show, count } = cfg;
    return poll(async () => {
      try {
        const [l, p] = await Promise.all([
          show === 'picture' ? null : optional(space.launches(count)),
          show === 'launches' ? null : optional(space.picture()),
        ]);
        launches = l;
        picture = p;
        off = l === null && p === null;
        error = '';
        now = new Date();
      } catch (err) {
        error = errorMessage(err);
        throw err;
      }
    }, POLL_MS);
  });

  const when = (d: Date) =>
    d.toLocaleString(intlLocale(), {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  const statusLabel = (s: string) => t(`space.status_${s}` as MessageKey);
</script>

{#if error && !launches && !picture}
  <p class="error">{error}</p>
{:else if off}
  <p class="muted">{t('space.switched_off')}</p>
{:else}
  {#if picture}
    <figure class="apod">
      {#if picture.image}
        <img src={picture.image} alt={picture.title} loading="lazy" />
      {:else}
        <p class="muted small">
          {t('space.video_today')}
          <a href={picture.link} target="_blank" rel="noopener noreferrer">{t('space.open')}</a>
        </p>
      {/if}
      <figcaption>
        <strong>{picture.title}</strong>
        {#if picture.copyright}<span class="muted small"> · © {picture.copyright}</span>{/if}
        <details>
          <summary class="small">{t('space.more')}</summary>
          <p class="small">{picture.explanation}</p>
        </details>
      </figcaption>
    </figure>
  {/if}
  {#if launches}
    <h4>{t('space.next_launches')}</h4>
    <ul class="launches">
      {#each launches as l (l.id)}
        <ListRow nowrap title={l.description || undefined}>
          {l.mission}
          {#snippet sub()}{l.rocket} · {l.provider}{l.location ? ` · ${l.location}` : ''}{/snippet}
          {#snippet value()}
            <span class="when">
              <span class="status {statusTone(l.status)}">{statusLabel(l.status)}</span>
              {#if l.webcastLive}<span class="live">{t('space.live')}</span>{/if}
              <time datetime={l.net.toISOString()} title={when(l.net)}>{countdown(l.net, now)}</time>
              <span class="small">{when(l.net)}</span>
            </span>
          {/snippet}
        </ListRow>
      {:else}
        <li class="muted">{t('space.no_launches')}</li>
      {/each}
    </ul>
  {/if}
{/if}

<style>
  .apod {
    margin: 0 0 var(--space-2);
  }
  img {
    display: block;
    width: 100%;
    max-height: 40cqh;
    object-fit: cover;
    border-radius: var(--radius-sm);
  }
  figcaption {
    margin-top: var(--space-1);
  }
  details p {
    margin: var(--space-1) 0 0;
  }
  h4 {
    margin: 0 0 var(--space-1);
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
  }
  .launches {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .when {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }
  time {
    font-weight: 600;
    color: var(--color-text);
  }
  .status {
    font-size: var(--fs-xs);
    padding: 0 var(--space-1);
    border-radius: var(--radius-sm);
    background: var(--color-surface-2);
  }
  .status.good {
    color: #2e7d32;
    background: color-mix(in srgb, #4caf50 18%, transparent);
  }
  .status.bad {
    color: var(--color-danger);
    background: color-mix(in srgb, var(--color-danger) 15%, transparent);
  }
  .live {
    font-size: var(--fs-xs);
    color: var(--color-danger);
    font-weight: 600;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
