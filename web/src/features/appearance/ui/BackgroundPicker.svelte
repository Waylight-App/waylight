<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { type MessageKey, t } from '../../../shared/i18n';
  import {
    backgroundCss,
    BUILTIN_BACKGROUNDS,
    BUILTIN_PREFIX,
    isUploadedBackground,
    isValidBackground,
  } from '../../../shared/theme/background';
  import Modal from '../../../shared/ui/Modal.svelte';
  import type { BackgroundLibrary, BackgroundScope } from '../domain/backgrounds';
  import type { AppearanceUseCases } from '../usecase/appearanceUseCases';

  /**
   * One picker for the global and the per-tab background (#277): built-in
   * gradients and colours, uploaded pictures and URLs, with a live preview.
   */
  interface Props {
    appearance: AppearanceUseCases;
    /** The global background as stored. */
    globalValue: string;
    /** Set when opened for a tab: its name and own background ("" = none). */
    tab?: { name: string; background: string };
    /** Persists the choice; clearTab removes the tab's own background too. */
    onsave: (value: string, scope: BackgroundScope, clearTab: boolean) => Promise<void>;
    onclose: () => void;
  }
  let { appearance, globalValue, tab, onsave, onclose }: Props = $props();

  // A tab with its own background opens on that tab; otherwise the global
  // background is what people usually want to change.
  // svelte-ignore state_referenced_locally
  let scope = $state<BackgroundScope>(tab?.background ? 'tab' : 'global');
  // svelte-ignore state_referenced_locally
  let value = $state(tab?.background ? tab.background : globalValue);
  let url = $state('');
  let clearTab = $state(true);
  let library = $state<BackgroundLibrary | null>(null);
  let error = $state('');
  let busy = $state(false);

  const builtinLabels: Record<string, MessageKey> = {
    dusk: 'appearance.background_builtin_dusk',
    dawn: 'appearance.background_builtin_dawn',
    forest: 'appearance.background_builtin_forest',
    ocean: 'appearance.background_builtin_ocean',
    aurora: 'appearance.background_builtin_aurora',
    sand: 'appearance.background_builtin_sand',
    slate: 'appearance.background_builtin_slate',
    paper: 'appearance.background_builtin_paper',
  };
  const builtinLabel = (name: string): string => {
    const key = builtinLabels[name];
    return key ? t(key) : name;
  };

  const preview = $derived(backgroundCss(value));
  const overridden = $derived(scope === 'global' && !!tab?.background);

  $effect(() => {
    appearance
      .backgrounds()
      .then((l) => (library = l))
      .catch((err) => (error = errorMessage(err)));
  });

  function switchScope(next: BackgroundScope) {
    scope = next;
    value = next === 'tab' ? (tab?.background ?? '') : globalValue;
    url = '';
  }

  function useUrl() {
    const v = url.trim();
    if (!isValidBackground(v) || isUploadedBackground(v) || v.startsWith(BUILTIN_PREFIX)) {
      error = t('appearance.background_invalid');
      return;
    }
    error = '';
    value = v;
  }

  async function upload(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file || !library) return;
    busy = true;
    error = '';
    try {
      const b = await appearance.uploadBackground(file, library.maxUploadBytes);
      library = { ...library, items: [b, ...library.items] };
      value = b.url;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function remove(id: string, picture: string) {
    if (!library || !confirm(t('appearance.background_delete_confirm'))) return;
    try {
      await appearance.deleteBackground(id);
      library = { ...library, items: library.items.filter((b) => b.id !== id) };
      if (value === picture) value = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function save() {
    busy = true;
    error = '';
    try {
      await onsave(value, scope, overridden && clearTab);
      onclose();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={t('appearance.background_title')} {onclose}>
  <div class="picker">
    {#if tab}
      <fieldset class="scope">
        <legend class="sr-only">{t('appearance.background_scope')}</legend>
        <label
          ><input
            type="radio"
            name="bg-scope"
            checked={scope === 'global'}
            onchange={() => switchScope('global')}
          />
          {t('appearance.background_scope_global')}</label
        >
        <label
          ><input
            type="radio"
            name="bg-scope"
            checked={scope === 'tab'}
            onchange={() => switchScope('tab')}
          />
          {t('appearance.background_scope_tab', { name: tab.name })}</label
        >
      </fieldset>
    {/if}

    <div class="preview" style:background-image={preview ?? 'none'} aria-live="polite">
      {#if !preview}<span
          >{scope === 'tab' ? t('appearance.background_none_tab') : t('appearance.background_default')}</span
        >{/if}
    </div>
    {#if overridden}
      <p class="notice" role="note">
        {t('appearance.background_overridden', { name: tab?.name ?? '' })}
        <label class="check"
          ><input type="checkbox" bind:checked={clearTab} /> {t('appearance.background_clear_tab')}</label
        >
      </p>
    {/if}

    <h3>{t('appearance.background_builtin')}</h3>
    <ul class="swatches">
      {#each BUILTIN_BACKGROUNDS as b (b.name)}
        <li>
          <button
            type="button"
            class="swatch"
            class:selected={value === BUILTIN_PREFIX + b.name}
            style:background-image={b.css}
            aria-label={builtinLabel(b.name)}
            aria-pressed={value === BUILTIN_PREFIX + b.name}
            onclick={() => (value = BUILTIN_PREFIX + b.name)}
          ></button>
        </li>
      {/each}
    </ul>

    <h3>{t('appearance.background_uploads')}</h3>
    <ul class="swatches">
      {#each library?.items ?? [] as b (b.id)}
        <li class="uploaded">
          <button
            type="button"
            class="swatch"
            class:selected={value === b.url}
            style:background-image={`url("${b.thumb}")`}
            aria-label={t('appearance.background_upload_pick', { w: b.width, h: b.height })}
            aria-pressed={value === b.url}
            onclick={() => (value = b.url)}
          ></button>
          <button
            type="button"
            class="btn btn-icon remove"
            aria-label={t('appearance.background_delete')}
            onclick={() => remove(b.id, b.url)}>✕</button
          >
        </li>
      {/each}
      <li>
        <label class="swatch add" class:busy>
          <span aria-hidden="true">+</span>
          <span class="sr-only">{t('appearance.background_upload')}</span>
          <input
            type="file"
            accept="image/jpeg,image/png,image/gif"
            onchange={upload}
            disabled={busy || (library ? library.items.length >= library.max : true)}
          />
        </label>
      </li>
    </ul>
    <p class="muted small">{t('appearance.background_upload_hint')}</p>

    <h3>{t('appearance.background_url')}</h3>
    <div class="url">
      <input class="input" type="url" bind:value={url} placeholder="https://…" maxlength="2048" />
      <button class="btn" type="button" onclick={useUrl} disabled={!url.trim()}
        >{t('appearance.background_use_url')}</button
      >
    </div>

    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <div class="actions">
      <button class="btn" type="button" onclick={() => (value = '')} disabled={value === ''}
        >{scope === 'tab' ? t('appearance.background_reset_tab') : t('appearance.background_reset')}</button
      >
      <button class="btn btn-primary" type="button" onclick={save} disabled={busy}>{t('common.save')}</button>
    </div>
  </div>
</Modal>

<style>
  .picker {
    display: grid;
    gap: var(--space-2);
    min-width: min(32rem, 100%);
  }
  .scope {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
    border: none;
    padding: 0;
    margin: 0;
  }
  .preview {
    height: 9rem;
    border-radius: var(--radius-small, 6px);
    border: 1px solid var(--color-border);
    background-color: var(--color-bg);
    background-size: cover;
    background-position: center;
    display: grid;
    place-items: center;
  }
  .notice {
    border-left: 3px solid var(--color-accent);
    padding-left: var(--space-2);
    margin: 0;
    display: grid;
    gap: var(--space-1);
  }
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  h3 {
    margin: var(--space-2) 0 0;
    font-size: 0.95em;
  }
  .swatches {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(4.5rem, 1fr));
    gap: var(--space-2);
  }
  .swatch {
    display: grid;
    place-items: center;
    width: 100%;
    aspect-ratio: 16 / 10;
    border: 2px solid var(--color-border);
    border-radius: var(--radius-small, 6px);
    background-size: cover;
    background-position: center;
    cursor: pointer;
    position: relative;
    padding: 0;
  }
  .swatch.selected {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px var(--color-accent);
  }
  .uploaded {
    position: relative;
  }
  .remove {
    position: absolute;
    top: 0;
    right: 0;
  }
  .add {
    font-size: 1.4rem;
    color: var(--color-text-muted, var(--color-text));
  }
  .add.busy {
    opacity: 0.6;
  }
  .add input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }
  .add:focus-within {
    outline: 2px solid var(--color-accent);
  }
  .url {
    display: flex;
    gap: var(--space-2);
  }
  .url .input {
    flex: 1;
  }
  .small {
    font-size: 0.85em;
    margin: 0;
  }
  .actions {
    display: flex;
    justify-content: space-between;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
  }
</style>
