<script lang="ts">
  import { t } from '../../../../shared/i18n';
  import { errorMessage } from '../../../../shared/api-client/http';
  import DesignGallery from '../../../appearance/ui/DesignGallery.svelte';
  import { previewSide, previewStyle, type ThemeGallery } from '../../../appearance/domain/themes';
  import type { StepProps } from './wizardContext';

  let { ctx, ctl }: StepProps = $props();

  let gallery = $state<ThemeGallery>({ themes: [], applied: '', canRestore: false });
  let error = $state('');
  let busy = $state(false);
  // svelte-ignore state_referenced_locally
  let accent = $state(ctx.personal?.accentColor ?? '');

  $effect(() => {
    ctx.design
      ?.gallery()
      .then((g) => (gallery = g))
      .catch((err) => (error = errorMessage(err)));
  });

  const chosen = $derived(gallery.themes.find((th) => th.id === gallery.applied));
  const dark = typeof matchMedia === 'function' && matchMedia('(prefers-color-scheme: dark)').matches;
  // The accent chosen here sits on top of the design, like in the settings.
  const preview = $derived(
    chosen
      ? `${previewStyle(chosen, previewSide(chosen, ctx.personal?.theme === 'dark' || (ctx.personal?.theme === 'system' && dark)))}${accent ? `; --color-accent: ${accent}` : ''}`
      : '',
  );

  async function select(id: string) {
    busy = true;
    error = '';
    try {
      const result = await ctx.design?.apply(id);
      accent = result?.accentColor ?? '';
      gallery = { ...gallery, applied: id };
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  // The shell hands each step a fresh control, so the first value is the one.
  // svelte-ignore state_referenced_locally
  ctl.save = async () => {
    if (accent !== (ctx.personal?.accentColor ?? '')) await ctx.savePersonal({ accentColor: accent });
  };

  const samples = [
    ['widget.weather.title', 'wizard.design.sample_weather'],
    ['widget.waste.title', 'wizard.design.sample_waste'],
    ['wizard.design.today', 'wizard.design.sample_today'],
    ['widget.shopping.title', 'wizard.design.sample_shopping'],
    ['widget.todos.title', 'wizard.design.sample_tasks'],
  ] as const;
</script>

<h1 class="wz-title wz-title-md" tabindex="-1">{t('wizard.design.title')}</h1>
<p class="wz-lead wz-lead-sm">{t('wizard.design.body')}</p>
<div class="look">
  <div>
    <DesignGallery
      themes={gallery.themes}
      selected={gallery.applied}
      filters={['all', 'light', 'dark']}
      compact
      disabled={busy}
      onselect={select}
    />
    <label class="accent">
      <span>{t('appearance.accent_color')}</span>
      <input type="color" bind:value={accent} />
      <small>{t('wizard.design.accent_hint')}</small>
    </label>
  </div>
  <div class="sample" style={preview} aria-hidden="true">
    {#each samples as [title, line] (title)}
      <div class="tile"><b>{t(title)}</b><span>{t(line)}</span></div>
    {/each}
  </div>
</div>
{#if error}<p class="wz-error" role="alert">{error}</p>{/if}

<style>
  .look {
    display: grid;
    grid-template-columns: minmax(0, 3fr) minmax(0, 2fr);
    gap: var(--space-6);
    align-items: start;
  }
  .accent {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-xs);
  }
  .accent small {
    opacity: 0.8;
  }
  /* The samples are drawn with the chosen design's own tokens. */
  .sample {
    display: grid;
    gap: var(--space-2);
    padding: var(--space-3);
    border-radius: var(--radius-md);
    background: var(--color-bg, transparent);
    color: var(--color-text);
  }
  .tile {
    display: grid;
    gap: var(--space-1);
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    font-size: var(--fs-xs);
  }
  .tile b {
    font-family: var(--font-display);
    font-size: var(--fs-sm);
    color: var(--color-accent-text, var(--color-accent));
  }
  @media (max-width: 40rem) {
    .look {
      grid-template-columns: 1fr;
    }
  }
</style>
