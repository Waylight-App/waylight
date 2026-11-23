<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import { formatBytes } from '../../../shared/ui/format';
  import {
    MAX_ALBUM_LENGTH,
    MAX_INTERVAL,
    MIN_INTERVAL,
    move,
    parseConfig,
    type Photo,
    type PhotoFit,
    type PhotoOverview,
    usedPercent,
  } from '../domain/photos';
  import type { PhotosUseCases } from '../usecase/photosUseCases';

  /** Settings of one photo frame: album, its pictures, interval and fit (#286). */
  interface Props {
    config: Record<string, unknown>;
    photos: PhotosUseCases;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, photos, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  const initial = parseConfig(config);
  let album = $state(initial.album);
  let interval = $state(initial.interval);
  let fit = $state<PhotoFit>(initial.fit);
  let list = $state<Photo[]>([]);
  let overview = $state<PhotoOverview | null>(null);
  let error = $state('');
  let uploading = $state<{ done: number; total: number } | null>(null);

  const albumName = $derived(album.trim() || initial.album);

  async function refresh() {
    try {
      [list, overview] = await Promise.all([photos.list(albumName), photos.overview()]);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    void albumName;
    void refresh();
  });

  async function upload(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const files = [...(input.files ?? [])];
    input.value = '';
    if (files.length === 0) return;
    error = '';
    uploading = { done: 0, total: files.length };
    const r = await photos.upload(albumName, files, (done) => (uploading = { done, total: files.length }));
    uploading = null;
    if (r.failed.length > 0) {
      const first = r.failed[0];
      error = t('photos.upload_failed', {
        n: r.failed.length,
        name: first?.name ?? '',
        reason: errorMessage(first?.error),
      });
    }
    await refresh();
  }

  async function shift(i: number, delta: number) {
    const next = move(list, i, delta);
    if (!next) return;
    const before = list;
    list = next;
    try {
      await photos.reorder(albumName, next);
    } catch (err) {
      list = before;
      error = errorMessage(err);
    }
  }

  async function remove(p: Photo) {
    if (!confirm(t('photos.confirm_delete'))) return;
    try {
      await photos.remove(p.id);
      await refresh();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...parseConfig({ album: albumName, interval, fit }) });
  }}
>
  <label class="field"
    ><span>{t('photos.album')}</span><input
      class="input"
      list="photo-albums"
      maxlength={MAX_ALBUM_LENGTH}
      bind:value={album}
    /></label
  >
  <datalist id="photo-albums">
    {#each overview?.albums ?? [] as a (a.name)}<option value={a.name}>{a.count}</option>{/each}
  </datalist>

  <div class="row">
    <label class="field"
      ><span>{t('photos.interval')}</span><input
        class="input"
        type="number"
        min={MIN_INTERVAL}
        max={MAX_INTERVAL}
        bind:value={interval}
      /></label
    >
    <label class="field"
      ><span>{t('photos.fit')}</span><select class="input" bind:value={fit}>
        <option value="contain">{t('photos.fit_contain')}</option>
        <option value="cover">{t('photos.fit_cover')}</option>
      </select></label
    >
  </div>

  <fieldset>
    <legend>{t('photos.pictures', { album: albumName })}</legend>
    <label class="btn file"
      >{uploading ? t('photos.uploading', uploading) : t('photos.upload')}<input
        type="file"
        accept="image/jpeg,image/png,image/gif"
        multiple
        disabled={uploading !== null}
        onchange={upload}
      /></label
    >
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <ol class="grid">
      {#each list as p, i (p.id)}
        <li>
          <img
            src={photos.imageUrl(p.id, true)}
            alt={t('photos.picture_of', { n: i + 1, total: list.length })}
            loading="lazy"
          />
          <span class="tools">
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('photos.move_earlier', { n: i + 1 })}
              disabled={i === 0}
              onclick={() => shift(i, -1)}>←</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('photos.move_later', { n: i + 1 })}
              disabled={i === list.length - 1}
              onclick={() => shift(i, 1)}>→</button
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('photos.delete', { n: i + 1 })}
              onclick={() => remove(p)}>✕</button
            >
          </span>
        </li>
      {/each}
    </ol>
    {#if overview}
      <p class="muted small">
        {t('photos.usage', {
          used: formatBytes(overview.usedBytes),
          quota: formatBytes(overview.quotaBytes),
          count: overview.count,
          max: overview.maxPhotos,
        })}
      </p>
      <progress max="100" value={usedPercent(overview)} aria-label={t('photos.usage_label')}></progress>
      <p class="muted small">{t('photos.backup_note')}</p>
    {/if}
  </fieldset>

  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  form {
    display: grid;
    gap: var(--space-2);
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .row .field {
    flex: 1 1 10rem;
  }
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    padding: var(--space-2);
    display: grid;
    gap: var(--space-2);
  }
  .file {
    position: relative;
    overflow: hidden;
    justify-self: start;
  }
  .file input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }
  .grid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(6.5rem, 1fr));
    gap: var(--space-2);
  }
  .grid li {
    display: grid;
    gap: var(--space-1);
  }
  .grid img {
    width: 100%;
    aspect-ratio: 4 / 3;
    object-fit: cover;
    border-radius: var(--radius-sm);
    background: var(--color-surface-2, transparent);
  }
  .tools {
    display: flex;
    justify-content: space-between;
  }
  progress {
    width: 100%;
  }
  .small {
    font-size: var(--fs-xs);
    margin: 0;
  }
</style>
