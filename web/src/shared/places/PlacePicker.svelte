<script lang="ts">
  import { t } from '../i18n';
  import { placeAt, type SavedPlace, type SavedPlaces } from './places';

  /**
   * Props of the saved-place choice in widget settings (#591). It shows
   * nothing while the places module is off, so the form around it (search,
   * coordinates) stays the way to enter a location.
   */
  interface Props {
    places: SavedPlaces;
    /** Id of the chosen place, '' for none. */
    value: string;
    /**
     * The location the form holds now. It preselects a saved place at the
     * same spot and can be saved as a new place when it matches none.
     */
    current?: { name: string; label: string; lat: number; lon: number } | null;
    onpick: (place: SavedPlace | null) => void;
  }
  let { places, value, current = null, onpick }: Props = $props();

  let list = $state<SavedPlace[]>([]);
  // False while the places module does not answer: nothing is offered.
  let available = $state(false);
  let saving = $state(false);
  let error = $state('');

  $effect(() => {
    places
      .list()
      .then((l) => {
        list = l;
        available = true;
        // A widget set up before places existed: preselect the place at its
        // coordinates, so it counts as using it.
        if (!value && current) {
          const match = placeAt(l, current.lat, current.lon);
          if (match) onpick(match);
        }
      })
      .catch(() => (available = false));
  });

  function change(e: Event) {
    const id = (e.currentTarget as HTMLSelectElement).value;
    onpick(list.find((p) => p.id === id) ?? null);
  }

  async function save() {
    if (!current) return;
    saving = true;
    error = '';
    try {
      const p = await places.add({
        name: current.name || current.label,
        label: current.label,
        lat: current.lat,
        lon: current.lon,
      });
      list = [...list, p];
      onpick(p);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      saving = false;
    }
  }
</script>

{#if available}
  {#if list.length}
    <label class="field">
      <span>{t('places.saved_place')}</span>
      <select class="input" {value} onchange={change}>
        <option value="">{t('places.other_location')}</option>
        {#each list as p (p.id)}
          <option value={p.id}>{p.label ? `${p.name} – ${p.label}` : p.name}</option>
        {/each}
      </select>
    </label>
  {/if}
  {#if !value && current && (current.name || current.label)}
    <button class="btn" type="button" disabled={saving} onclick={save}>{t('places.save_as_place')}</button>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
{/if}
