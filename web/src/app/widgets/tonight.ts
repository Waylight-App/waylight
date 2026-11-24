import { t } from '../../shared/i18n';
import { compactSource } from '../../features/compact/domain/compact';
import { parseWeatherConfig } from '../../features/weather/domain/weather';
import { tonightFor } from '../tonight/weatherSource';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The Tonight card: clock and weather in one lit card. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'tonight',
      category: 'everyday',
      keywords: ['clock', 'weather', 'time', 'forecast', 'evening', 'uhr', 'wetter', 'zeit', 'abend'],
      get title() {
        return t('widget.tonight.title');
      },
      get description() {
        return t('widget.tonight.description');
      },
      defaultSize: { w: 6, h: 4 },
    },
    icon: 'sun',
    // The compact state is the weather's, for the same place; small and strip
    // tiles draw it with the time in a view of their own (compactView).
    compact: (config) => {
      const c = parseWeatherConfig(config);
      return c ? compactSource('weather', { lat: String(c.lat), lon: String(c.lon) }) : undefined;
    },
    compactExample: () => t('compact.example.tonight'),
    chromeless: true,
    view: lazy(
      () => import('../tonight/TonightView.svelte'),
      (config, c) => ({ config, tonight: tonightFor(c.weather) }),
    ),
    compactView: lazy(
      () => import('../../features/tonight/ui/TonightCompact.svelte'),
      (config, c) => ({ config, tonight: tonightFor(c.weather) }),
    ),
    settings: lazy(
      () => import('../../features/weather/ui/WeatherSettings.svelte'),
      (config, c) => ({ config, weather: c.weather, mode: 'radar' as const, places: c.places }),
    ),
  },
];
