<script lang="ts">
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { Appearance } from '../domain/appearance';
  import type { ThemeGallery } from '../domain/themes';
  import DesignGallery from './DesignGallery.svelte';
  import type { AppearanceUseCases } from '../usecase/appearanceUseCases';

  /** Theme gallery (#111): ready-made themes to apply, import and share. */
  interface Props {
    appearance: AppearanceUseCases;
    /** Called with the preferences after applying or restoring. */
    onapply: (a: Appearance) => void;
  }
  let { appearance, onapply }: Props = $props();

  let gallery = $state<ThemeGallery>({ themes: [], applied: '', canRestore: false });
  let error = $state('');
  let busy = $state(false);
  let picker = $state<HTMLInputElement>();

  const load = async () => (gallery = await appearance.gallery());
  $effect(() => {
    load().catch((err) => (error = errorMessage(err)));
  });

  async function run(action: () => Promise<Appearance | unknown>, apply = false) {
    error = '';
    busy = true;
    try {
      const result = await action();
      if (apply) onapply(result as Appearance);
      await load();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function importFile() {
    const file = picker?.files?.[0];
    if (!file) return;
    void run(() => appearance.importTheme(file)).then(() => picker && (picker.value = ''));
  }
</script>

<section aria-labelledby="theme-gallery">
  <h3 id="theme-gallery">{t('appearance.gallery')}</h3>
  <p class="muted small">{t('appearance.gallery_hint')}</p>
  <DesignGallery
    themes={gallery.themes}
    selected={gallery.applied}
    filters={['all', 'light', 'dark', 'own']}
    disabled={busy}
    onselect={(id) => run(() => appearance.applyTheme(id), true)}
  >
    {#snippet extra()}
      <label class="btn import"
        >{t('appearance.import_theme')}<input
          class="file"
          bind:this={picker}
          type="file"
          accept="application/json,.json"
          disabled={busy}
          onchange={importFile}
        /></label
      >
    {/snippet}
    {#snippet detail(theme)}
      {#if theme.author}<span class="muted">{theme.author}</span>{/if}
      <a
        class="btn btn-icon"
        href={appearance.exportUrl(theme.id)}
        download
        aria-label={t('appearance.export_theme', { name: theme.name })}><Icon name="download" size={16} /></a
      >
      {#if !theme.builtIn}
        <button
          class="btn btn-icon"
          type="button"
          disabled={busy}
          aria-label={t('appearance.delete_theme', { name: theme.name })}
          onclick={() => run(() => appearance.deleteTheme(theme.id))}>✕</button
        >
      {/if}
    {/snippet}
  </DesignGallery>
  {#if gallery.canRestore}
    <button
      class="btn"
      type="button"
      disabled={busy}
      onclick={() => run(() => appearance.restoreStyle(), true)}>{t('appearance.restore_own')}</button
    >
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
  h3 {
    margin: var(--space-4) 0 var(--space-1);
    font-size: var(--fs-sm);
  }
  .muted {
    color: var(--color-text-subtle);
  }
  .small {
    font-size: var(--fs-xs);
  }
  .import {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    min-height: 6rem;
    border: 2px dashed var(--color-border);
    border-radius: var(--radius-md);
    cursor: pointer;
  }
  .file {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
  }
  .import:has(:focus-visible) {
    outline: 2px solid var(--color-accent-text);
  }
</style>
