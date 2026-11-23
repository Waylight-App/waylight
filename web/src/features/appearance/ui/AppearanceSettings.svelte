<script lang="ts">
  import { backgroundCss } from '../../../shared/theme/background';
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Appearance, parseTextSize, TEXT_SIZES } from '../domain/appearance';
  import type { AppearanceUseCases } from '../usecase/appearanceUseCases';
  import ThemeGallery from './ThemeGallery.svelte';

  /** Props of the appearance settings section. */
  interface Props {
    appearance: AppearanceUseCases;
    current: Appearance;
    /** Called with the saved preferences so the app can apply them. */
    onchange: (a: Appearance) => void;
    /** Shows the advanced fields (#602); custom CSS that is set stays visible anyway. */
    advanced?: boolean;
  }
  let { appearance, current, onchange, advanced = true }: Props = $props();

  // svelte-ignore state_referenced_locally
  let form = $state<Appearance>({ ...current, textSize: parseTextSize(current.textSize) });
  let error = $state('');
  let saved = $state(false);
  let picking = $state(false);
  const loadPicker = () => import('./BackgroundPicker.svelte');

  // The picker saves right away, like on the dashboard; the rest of the
  // form keeps waiting for its own Save button.
  async function saveBackground(value: string) {
    const result = await appearance.setBackground(value);
    form.backgroundImage = result.backgroundImage;
    onchange({ ...current, backgroundImage: result.backgroundImage });
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    saved = false;
    try {
      const result = await appearance.save(form);
      onchange(result);
      saved = true;
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={submit}>
  <label class="field">
    <span>{t('appearance.greeting_name')}</span>
    <input class="input" bind:value={form.greetingName} maxlength="40" autocomplete="given-name" />
  </label>
  <label class="check">
    <input type="checkbox" bind:checked={form.greetingMoment} />
    <span>{t('appearance.greeting_moment')}</span>
  </label>
  <label class="field">
    <span>{t('appearance.profile')}</span>
    <select class="input" bind:value={form.profile}>
      <option value="everyday">{t('auth.profile_everyday')} – {t('auth.profile_everyday_hint')}</option>
      <option value="homelab">{t('auth.profile_homelab')}</option>
    </select>
    <span class="muted small">{t('appearance.profile_hint')}</span>
  </label>
  <label class="field" id="settings-launcher">
    <span>{t('appearance.launcher')}</span>
    <select class="input" bind:value={form.launcher}>
      <option value="both">{t('appearance.launcher_both')}</option>
      <option value="services">{t('appearance.launcher_services')}</option>
      <option value="bookmarks">{t('appearance.launcher_bookmarks')}</option>
    </select>
  </label>
  <label class="field">
    <span>{t('appearance.language')}</span>
    <select class="input" bind:value={form.language}>
      <option value="">{t('appearance.language_browser')}</option>
      <option value="en">English</option>
      <option value="de">Deutsch</option>
    </select>
  </label>
  <label class="field">
    <span>{t('appearance.theme')}</span>
    <select class="input" bind:value={form.theme} disabled={form.themeLocked}>
      <option value="system">{t('appearance.system')}</option>
      <option value="light">{t('appearance.light')}</option>
      <option value="dark">{t('appearance.dark')}</option>
    </select>
  </label>
  <label class="field">
    <span>{t('appearance.accent_color')}</span>
    <span class="color">
      <input
        type="color"
        value={form.accentColor || '#3365e8'}
        disabled={form.themeLocked}
        oninput={(e) => (form.accentColor = e.currentTarget.value)}
      />
      <input
        class="input"
        bind:value={form.accentColor}
        placeholder={t('appearance.default')}
        disabled={form.themeLocked}
      />
    </span>
  </label>
  {#if form.themeLocked}<p class="muted small">{t('appearance.theme_locked')}</p>{/if}
  <div class="field">
    <span>{t('appearance.background_image')}</span>
    <span class="bg-row">
      <span class="bg-preview" style:background-image={backgroundCss(form.backgroundImage) ?? 'none'}></span>
      <button class="btn" type="button" onclick={() => (picking = true)}
        >{t('appearance.background_choose')}</button
      >
    </span>
  </div>
  <fieldset class="access" id="settings-accessibility">
    <legend>{t('appearance.accessibility')}</legend>
    <label class="field">
      <span>{t('appearance.text_size')}</span>
      <select class="input" bind:value={form.textSize}>
        {#each TEXT_SIZES as size (size)}
          <option value={size}>{t(`appearance.text_size_${size}`)}</option>
        {/each}
      </select>
    </label>
    <label class="check">
      <input type="checkbox" bind:checked={form.highContrast} aria-describedby="high-contrast-hint" />
      <span>{t('appearance.high_contrast')}</span>
    </label>
    <p class="muted small hint" id="high-contrast-hint">{t('appearance.high_contrast_hint')}</p>
    <label class="check">
      <input type="checkbox" bind:checked={form.simpleView} aria-describedby="simple-view-hint" />
      <span>{t('appearance.simple_view')}</span>
    </label>
    <p class="muted small hint" id="simple-view-hint">{t('appearance.simple_view_hint')}</p>
  </fieldset>
  {#if advanced || current.customCss.trim() !== ''}
    <label class="field">
      <span>{t('appearance.custom_css_override_variables_such_as')}</span>
      <textarea class="input css" bind:value={form.customCss} rows="6" spellcheck="false"></textarea>
    </label>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if saved}<p class="muted">Saved.</p>{/if}
  <button class="btn btn-primary" type="submit">{t('appearance.save_appearance')}</button>
</form>

<ThemeGallery
  {appearance}
  onapply={(a) => {
    form = { ...a };
    saved = false;
    onchange(a);
  }}
/>

{#if picking}
  {#await loadPicker() then { default: BackgroundPicker }}
    <BackgroundPicker
      {appearance}
      globalValue={form.backgroundImage}
      onsave={(value) => saveBackground(value)}
      onclose={() => (picking = false)}
    />
  {/await}
{/if}

<style>
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
    margin-bottom: var(--space-2);
  }
  .bg-row {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  .bg-preview {
    width: 5rem;
    aspect-ratio: 16 / 10;
    border-radius: var(--radius-sm);
    border: 1px solid var(--color-border);
    background-color: var(--color-bg);
    background-size: cover;
    background-position: center;
  }
  .color {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  .css {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
  }
  .small {
    font-size: var(--fs-xs);
    margin: 0;
  }
  .access {
    margin: 0 0 var(--space-3);
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
  }
  .access legend {
    padding: 0 var(--space-1);
    font-weight: 600;
  }
  .hint {
    margin: 0 0 var(--space-2);
  }
</style>
