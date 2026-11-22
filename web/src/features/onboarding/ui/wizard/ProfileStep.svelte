<script lang="ts">
  import { locale, t } from '../../../../shared/i18n';
  import { errorMessage } from '../../../../shared/api-client/http';
  import { browserLocation } from '../../../../shared/places/places';
  import type { Place } from '../../domain/profile';
  import type { StepProps } from './wizardContext';

  let { ctx, session, ctl }: StepProps = $props();

  type Found = Omit<Place, 'id'>;

  // svelte-ignore state_referenced_locally
  let name = $state(ctx.personal?.greetingName ?? '');
  // The chips show the language in use; nothing is saved unless one is picked.
  // svelte-ignore state_referenced_locally
  let language = $state<'en' | 'de'>(ctx.personal?.language || locale());
  let languagePicked = $state(false);
  // svelte-ignore state_referenced_locally
  let query = $state(session.place?.label ?? '');
  let results = $state<Found[]>([]);
  let chosen = $state<Found | undefined>();
  let error = $state('');
  let timer: ReturnType<typeof setTimeout> | undefined;

  const languages = [
    { id: 'de', label: 'Deutsch' },
    { id: 'en', label: 'English' },
  ] as const;

  function pick(id: 'de' | 'en') {
    language = id;
    languagePicked = true;
    ctx.preview({ language: id });
  }

  function typed() {
    chosen = undefined;
    error = '';
    clearTimeout(timer);
    if (!ctx.places?.search || query.trim().length < 3) {
      results = [];
      return;
    }
    timer = setTimeout(() => void find(), 400);
  }

  async function find(): Promise<Found[]> {
    try {
      results = (await ctx.places?.search?.(query.trim())) ?? [];
    } catch (err) {
      error = errorMessage(err);
      results = [];
    }
    return results;
  }

  function choose(place: Found) {
    chosen = place;
    query = place.label;
    results = [];
  }

  async function locate() {
    error = '';
    try {
      choose({ name: '', label: t('first_start.place.here'), ...(await browserLocation()) });
    } catch {
      error = t('templates.location_denied');
    }
  }

  // The shell hands each step a fresh control, so the first value is the one.
  // svelte-ignore state_referenced_locally
  ctl.save = async () => {
    const patch: Parameters<typeof ctx.savePersonal>[0] = {};
    if (name.trim() !== (ctx.personal?.greetingName ?? '')) patch.greetingName = name.trim();
    if (languagePicked && language !== ctx.personal?.language) patch.language = language;
    if (Object.keys(patch).length > 0) await ctx.savePersonal(patch);
    if (!ctx.places) return;
    clearTimeout(timer);
    // Text without a pick takes the first match, so a typed town is not lost.
    let place = chosen;
    const typedOnly = query.trim() !== '' && query.trim() !== session.place?.label;
    if (!place && typedOnly) {
      place = (await find())[0];
      if (!place) throw new Error(t('templates.no_results'));
    }
    if (place) session.place = await ctx.places.save({ ...place, name: t('first_start.place.name') });
  };
  // The shell hands each step a fresh control, so the first value is the one.
  // svelte-ignore state_referenced_locally
  ctl.skip = () => ctx.preview(null);
</script>

<h1 class="wz-title" tabindex="-1">{t('wizard.profile.title')}</h1>
<p class="wz-lead">{t('wizard.profile.body')}</p>
<div class="wz-form">
  {#if ctx.personal}
    <div class="wz-field">
      <label for="wz-name">{t('wizard.profile.name')}</label>
      <input id="wz-name" class="wz-input" bind:value={name} maxlength="40" autocomplete="given-name" />
    </div>
    <fieldset class="wz-field">
      <legend>{t('wizard.profile.language')}</legend>
      <div class="wz-chips">
        {#each languages as l (l.id)}
          <label class="wz-chip" class:on={language === l.id}>
            <input
              type="radio"
              name="wz-language"
              value={l.id}
              checked={language === l.id}
              onclick={() => pick(l.id)}
            />
            {l.label}
          </label>
        {/each}
      </div>
    </fieldset>
  {/if}
  {#if ctx.places}
    <div class="wz-field">
      {#if ctx.places.search}
        <label for="wz-place">{t('wizard.profile.place')}</label>
        <input
          id="wz-place"
          class="wz-input"
          bind:value={query}
          oninput={typed}
          autocomplete="off"
          enterkeyhint="search"
        />
        {#if results.length > 0}
          <ul class="wz-results" aria-label={t('first_start.place.results')}>
            {#each results.slice(0, 4) as r (`${r.lat},${r.lon}`)}
              <li><button type="button" onclick={() => choose(r)}>{r.label}</button></li>
            {/each}
          </ul>
        {/if}
      {:else}
        <span class="wz-label">{t('wizard.profile.place')}</span>
        {#if chosen}<p class="wz-hint" role="status">
            {t('first_start.place.chosen', { place: chosen.label })}
          </p>{/if}
      {/if}
      <button class="wz-sub" type="button" onclick={locate}>
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"><circle cx="12" cy="12" r="3" /><path d="M12 2v3M12 19v3M2 12h3M19 12h3" /></svg
        >{t('wizard.profile.locate')}
      </button>
    </div>
  {/if}
  {#if error}<p class="wz-error" role="alert">{error}</p>{/if}
</div>
