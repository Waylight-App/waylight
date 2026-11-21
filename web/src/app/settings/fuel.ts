import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'fuel',
  label: 'settings.section.fuel',
  module: 'fuel',
  load: () => import('../../features/fuel').then((m) => m.FuelSettings),
  props: (c, ctx) => ({
    fuel: c.fuel,
    places: ctx.modules.has('places') ? c.places : undefined,
    searchPlaces: ctx.modules.has('weather') ? (q: string) => c.weather.searchPlaces(q) : undefined,
  }),
});
