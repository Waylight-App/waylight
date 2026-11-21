import { defineSettingsSection } from '../settingsKit';
import { definitionFor } from '../widgets';
import { placeLabel } from '../../features/weather/domain/weather';

// Connections & places (#591): places entered once for every widget that
// needs a location. The settings start page (#600) shows it as the area
// "Connections & services".
export const section = defineSettingsSection({
  id: 'connections',
  label: 'settings.section.connections',
  module: 'places',
  load: () => import('../../features/places').then((m) => m.PlacesManager),
  props: (c, ctx) => ({
    places: c.places,
    // The weather provider is the geocoder; without it a place comes from
    // the browser's location only.
    search: ctx.modules.has('weather')
      ? async (q: string) =>
          (await c.weather.searchPlaces(q)).map((p) => ({
            name: p.name,
            label: placeLabel(p),
            lat: p.lat,
            lon: p.lon,
          }))
      : undefined,
    widgetName: (kind: string) => definitionFor(kind)?.type.title ?? kind,
    tabs: () => c.dashboard.load(),
  }),
});
