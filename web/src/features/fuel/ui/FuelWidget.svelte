<script lang="ts">
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { type ChoiceLoader, savedConnections } from '../../../shared/connection/connection';
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import { fuelSetup } from './fuelSetup';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import {
    type PlaceResult,
    type Favourite,
    formatPrice,
    fuelName,
    type FuelOverview,
    isFavourite,
    parseWidgetConfig,
    type Station,
    toggleFavourite,
  } from '../domain/fuel';
  import type { FuelUseCases } from '../usecase/fuelUseCases';

  /** Props of the fuel price widget (#53). */
  interface Props {
    fuel: FuelUseCases;
    config: Record<string, unknown>;
    /** Place search of the weather module for the setup; without it the settings page sets fuel up. */
    searchPlaces?: (query: string) => Promise<PlaceResult[]>;
    /** Saved connections to offer in the setup, and storing the choice (#709). */
    connections?: ChoiceLoader;
    onsave?: (config: Record<string, unknown>) => Promise<void>;
  }
  let { fuel, config, searchPlaces, connections, onsave }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseWidgetConfig(config));

  const POLL_MS = 5 * 60_000;
  let overview = $state<FuelOverview | null>(null);
  let favourites = $state<Favourite[]>([]);
  let error = $state('');
  let unconfigured = $state(false);
  const list = $derived(
    overview ? (cfg.view === 'favourites' ? overview.favourites : overview.cheapest) : [],
  );

  const load = () =>
    Promise.all([fuel.overview(cfg.connectionId), fuel.settings()])
      .then(([o, s]) => ((overview = o), (favourites = s.favourites), (error = ''), (unconfigured = false)))
      .catch((err) => {
        unconfigured = err instanceof ApiError && err.status === 404;
        error = errorMessage(err);
        throw err;
      });

  $effect(() => poll(load, POLL_MS));
  // svelte-ignore state_referenced_locally
  const setup = searchPlaces ? fuelSetup(fuel, searchPlaces) : null;

  async function star(s: Station) {
    try {
      favourites = (await fuel.setFavourites(toggleFavourite(favourites, s))).favourites;
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

{#if unconfigured && setup}
  <SetupInPlace
    text={t('fuel.connect_under_settings')}
    descriptor={setup}
    instance
    connections={savedConnections(['tankerkoenig'], connections, config, onsave)}
    ondone={() => void load().catch(() => undefined)}
  />
{:else if unconfigured}
  <p class="muted">{t('fuel.connect_under_settings')}</p>
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  <p class="muted small">{fuelName(overview.fuel)}{overview.place ? ` · ${overview.place}` : ''}</p>
  <ul class="stations">
    {#each list as s (s.id)}
      {@const p = formatPrice(s.price, intlLocale())}
      <ListRow nowrap title={s.address} style={s.open ? '' : 'opacity: 0.6'}>
        {#snippet lead()}
          <span class="price"
            >{#if p}{p.main}<sup>{p.tenth}</sup>{:else}–{/if}</span
          >
        {/snippet}
        {s.brand || s.name}
        {#snippet sub()}{s.open ? '' : `${t('fuel.closed')} · `}{s.distanceKm > 0
            ? `${s.distanceKm} km`
            : s.address}{/snippet}
        {#snippet trail()}
          <button
            class="btn btn-icon"
            type="button"
            aria-pressed={isFavourite(favourites, s.id)}
            aria-label={t('fuel.favourite', { name: s.brand || s.name })}
            onclick={() => star(s)}>{isFavourite(favourites, s.id) ? '★' : '☆'}</button
          >
        {/snippet}
      </ListRow>
    {:else}
      <li class="muted">{cfg.view === 'favourites' ? t('fuel.no_favourites') : t('fuel.no_stations')}</li>
    {/each}
  </ul>
  {#if error}<p class="error small">{error}</p>{/if}
  <p class="muted tiny">{t('fuel.attribution')}</p>
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .stations {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  /* The tenth of a cent keeps the guide's smallest size instead of the browser's 12px. */
  sup {
    font-size: var(--fs-xs);
  }
  .price {
    font-weight: 600;
    min-width: 3.6em;
  }
  .small {
    font-size: var(--fs-xs);
    margin: 0 0 var(--space-1);
  }
  .tiny {
    font-size: var(--fs-xs);
    margin: var(--space-1) 0 0;
  }
</style>
