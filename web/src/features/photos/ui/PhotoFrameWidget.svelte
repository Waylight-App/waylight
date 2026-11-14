<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import { nextIndex, parseConfig, type Photo } from '../domain/photos';
  import type { PhotosUseCases } from '../usecase/photosUseCases';

  /** Props of the photo frame widget (#286). */
  interface Props {
    photos: PhotosUseCases;
    config: Record<string, unknown>;
  }
  let { photos, config }: Props = $props();
  const poll = usePolling();

  const cfg = $derived(parseConfig(config));
  let list = $state<Photo[]>([]);
  let index = $state(0);
  let error = $state('');
  let loaded = $state(false);

  async function load() {
    try {
      list = await photos.list(cfg.album);
      if (index >= list.length) index = 0;
      error = '';
    } catch (err) {
      error = errorMessage(err);
      throw err;
    } finally {
      loaded = true;
    }
  }

  // New uploads show up within a few minutes; the settings reload on save anyway.
  $effect(() => poll(load, 5 * 60_000));

  // Rotation pauses while the tab is hidden, so a background tab neither
  // downloads pictures nor jumps ahead.
  $effect(() => {
    const count = list.length;
    const ms = cfg.interval * 1000;
    if (count < 2) return;
    const timer = setInterval(() => {
      if (document.visibilityState === 'visible') index = nextIndex(index, count);
    }, ms);
    return () => clearInterval(timer);
  });

  const current = $derived(list[index]);
  // Preloading the next picture keeps the fade free of a blank frame.
  const upcoming = $derived(list.length > 1 ? list[nextIndex(index, list.length)] : undefined);
</script>

{#if error}<p class="error">{error}</p>{/if}
{#if current}
  <div class="frame" class:cover={cfg.fit === 'cover'}>
    {#key current.id}
      <img
        src={photos.imageUrl(current.id)}
        alt={t('photos.picture_of', { n: index + 1, total: list.length })}
      />
    {/key}
    {#if upcoming}<link rel="preload" as="image" href={photos.imageUrl(upcoming.id)} />{/if}
  </div>
{:else if loaded && !error}
  <p class="muted empty">{t('photos.empty', { album: cfg.album })}</p>
{/if}

<style>
  .frame {
    position: relative;
    height: 100%;
    min-height: 6rem;
    overflow: hidden;
    border-radius: var(--radius-sm, 6px);
  }
  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: contain;
    animation: fade 0.8s ease-out;
  }
  .cover img {
    object-fit: cover;
  }
  @keyframes fade {
    from {
      opacity: 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    img {
      animation: none;
    }
  }
  .empty {
    margin: 0;
  }
</style>
