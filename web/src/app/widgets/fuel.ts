import { t } from '../../shared/i18n';
import { choicesOf, lazy, type WidgetDefinition } from '../widgetKit';

/** The fuel widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'fuel',
      category: 'everyday',
      keywords: ['petrol', 'gas', 'diesel', 'benzin', 'tanken', 'sprit', 'tankerkönig'],
      get title() {
        return t('widget.fuel.title');
      },
      get description() {
        return t('widget.fuel.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'gauge',
    view: lazy(
      () => import('../../features/fuel/ui/FuelWidget.svelte'),
      // The setup assistant finds the place with the weather module's search.
      (config, c) => ({
        config,
        fuel: c.fuel,
        searchPlaces: (q: string) => c.weather.searchPlaces(q),
        connections: choicesOf(c),
      }),
    ),
    settings: lazy(
      () => import('../../features/fuel/ui/FuelWidgetSettings.svelte'),
      (config, c) => ({ config, connections: choicesOf(c) }),
    ),
  },
];
