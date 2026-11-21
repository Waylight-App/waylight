<script lang="ts">
  import { errorMessage } from '../../shared/api-client/http';
  import { t } from '../../shared/i18n';
  import Modal from '../../shared/ui/Modal.svelte';
  import { copyToClipboard } from '../../shared/ui/clipboard';
  import { qrDataUrl } from '../../shared/ui/qr';
  import { browserLocation } from '../../shared/places/places';
  import { kioskLink } from '../../features/auth/domain/auth';
  import type { Container } from '../container';
  import { defaultChoice, kioskInput, ready, WALL_TEMPLATE, type WallTabletChoice } from './wallTablet';

  /**
   * "Set up a wall tablet" (#680): one form that creates the tab from the
   * Wall tablet template (or takes existing tabs) and the kiosk display
   * with large text, night mode and rotation, then shows the link and its
   * QR code for the tablet's browser or the app's wall display mode.
   */
  interface Props {
    container: Container;
    tabs: { id: string; name: string }[];
    /** Place search for the template's weather; omitted without the weather module. */
    searchPlaces?: (query: string) => Promise<{ label: string; lat: number; lon: number }[]>;
    /** A display was created; the kiosk list reloads. */
    oncreated: () => void;
    onclose: () => void;
  }
  let { container, tabs, searchPlaces, oncreated, onclose }: Props = $props();

  let choice = $state<WallTabletChoice>(defaultChoice());
  let place = $state<{ label: string; lat: number; lon: number } | null>(null);
  let query = $state('');
  let results = $state<{ label: string; lat: number; lon: number }[]>([]);
  let busy = $state(false);
  let error = $state('');
  let link = $state('');
  let copied = $state(false);
  // A tab the template created before the display failed (e.g. the name was
  // taken); a second try uses it instead of creating another one.
  let createdTab = $state<string | null>(null);

  const ROTATIONS = [15, 30, 60, 120, 300];

  async function search() {
    if (!searchPlaces || !query.trim()) return;
    error = '';
    try {
      results = await searchPlaces(query.trim());
      if (results.length === 0) error = t('templates.no_results');
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function here() {
    error = '';
    try {
      const at = await browserLocation();
      place = { label: t('tablet.my_location'), ...at };
    } catch {
      error = t('tablet.location_failed');
    }
  }

  async function create(e: SubmitEvent) {
    e.preventDefault();
    if (!ready(choice) || busy) return;
    busy = true;
    error = '';
    try {
      let tab: string | undefined;
      if (choice.source === 'template') {
        createdTab ??= (
          await container.dashboard.createTabFromBuiltin(
            WALL_TEMPLATE,
            t('templates.wall'),
            place ? { lat: place.lat, lon: place.lon, label: place.label } : undefined,
          )
        ).id;
        tab = createdTab;
      }
      const { token } = await container.auth.createKiosk(kioskInput(choice, tab));
      link = kioskLink(location.origin, token);
      oncreated();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  const toggled = (list: string[], id: string, on: boolean) =>
    on ? [...list, id] : list.filter((x) => x !== id);
</script>

<Modal title={t('tablet.title')} {onclose}>
  {#if link}
    <div class="done" role="status">
      <p><strong>{t('tablet.ready', { name: choice.name })}</strong></p>
      <img src={qrDataUrl(link, 5)} alt={t('auth.kiosk_qr')} width="205" height="205" />
      <ol>
        <li>{t('tablet.step_browser')}</li>
        <li>{t('tablet.step_app')}</li>
        <li>{t('tablet.step_fullscreen')}</li>
      </ol>
      <code>{link}</code>
      <p class="muted">{t('auth.kiosk_link_once')}</p>
      <p class="row">
        <button class="btn" type="button" onclick={async () => (copied = await copyToClipboard(link))}
          >{copied ? t('auth.kiosk_copied') : t('auth.kiosk_copy')}</button
        >
        <button class="btn btn-primary" type="button" onclick={onclose}>{t('auth.kiosk_done')}</button>
      </p>
    </div>
  {:else}
    <form onsubmit={create}>
      <p class="muted">{t('tablet.intro')}</p>
      <label class="field"
        ><span>{t('tablet.name')}</span>
        <input class="input" bind:value={choice.name} required maxlength="32" autocomplete="off" /></label
      >

      <fieldset>
        <legend>{t('tablet.what')}</legend>
        <label class="choice"
          ><input type="radio" bind:group={choice.source} value="template" />
          {t('tablet.from_template')}</label
        >
        <label class="choice"
          ><input type="radio" bind:group={choice.source} value="tabs" disabled={tabs.length === 0} />
          {t('tablet.existing_tabs')}</label
        >
        {#if choice.source === 'tabs'}
          <div class="tabs">
            {#each tabs as tab (tab.id)}
              <label class="choice"
                ><input
                  type="checkbox"
                  checked={choice.tabs.includes(tab.id)}
                  onchange={(e) => (choice.tabs = toggled(choice.tabs, tab.id, e.currentTarget.checked))}
                />
                {tab.name}</label
              >
            {/each}
          </div>
        {:else if !createdTab}
          <p class="muted small">
            {place ? t('tablet.weather_for', { place: place.label }) : t('tablet.weather_hint')}
          </p>
          <div class="row">
            <button class="btn" type="button" onclick={here}>{t('tablet.use_location')}</button>
          </div>
        {/if}
      </fieldset>
      {#if choice.source === 'template' && !createdTab && searchPlaces}
        <!-- Enter in the search field searches; it must not create the display. -->
        <div class="search">
          <div class="row">
            <input
              class="input"
              bind:value={query}
              placeholder={t('tablet.search_place')}
              aria-label={t('tablet.search_place')}
              onkeydown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  void search();
                }
              }}
            />
            <button class="btn" type="button" onclick={search}>{t('tablet.search')}</button>
          </div>
          {#if results.length}
            <ul class="results">
              {#each results as r (`${r.lat},${r.lon}`)}
                <li>
                  <button class="btn" type="button" onclick={() => ((place = r), (results = []))}
                    >{r.label}</button
                  >
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      {/if}

      <label class="choice"
        ><input type="checkbox" bind:checked={choice.largeText} /> {t('tablet.large_text')}</label
      >
      <label class="choice"><input type="checkbox" bind:checked={choice.night} /> {t('tablet.night')}</label>
      {#if choice.night}
        <div class="row">
          <label class="field"
            ><span>{t('auth.kiosk_night_from')}</span><input
              class="input"
              type="time"
              bind:value={choice.nightFrom}
              required
            /></label
          >
          <label class="field"
            ><span>{t('auth.kiosk_night_to')}</span><input
              class="input"
              type="time"
              bind:value={choice.nightTo}
              required
            /></label
          >
        </div>
      {/if}
      {#if choice.source === 'tabs' && choice.tabs.length > 1}
        <label class="field"
          ><span>{t('auth.kiosk_rotation')}</span>
          <select class="input" bind:value={choice.rotateSeconds}>
            {#each ROTATIONS as s (s)}<option value={s}>{t('auth.kiosk_rotates', { n: s })}</option>{/each}
          </select>
        </label>
      {/if}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <button class="btn btn-primary" type="submit" disabled={!ready(choice) || busy}
        >{t('tablet.create')}</button
      >
    </form>
  {/if}
</Modal>

<style>
  fieldset {
    border: none;
    padding: 0;
    margin: 0 0 var(--space-2);
  }
  legend {
    font-weight: 600;
    margin-bottom: var(--space-1);
  }
  .choice {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 44px;
  }
  .tabs {
    padding-left: var(--space-4);
  }
  .row {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    align-items: end;
  }
  .search {
    margin-bottom: var(--space-2);
  }
  .results {
    list-style: none;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .done code {
    display: block;
    word-break: break-all;
    margin: var(--space-2) 0;
  }
  .small {
    font-size: 0.85em;
  }
</style>
