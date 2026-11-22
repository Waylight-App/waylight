import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The meter readings widget (#689). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'meters',
      category: 'everyday',
      keywords: [
        'meters',
        'meter',
        'readings',
        'consumption',
        'electricity',
        'gas',
        'water',
        'zähler',
        'zählerstand',
        'ablesung',
        'verbrauch',
        'strom',
        'wasser',
      ],
      get title() {
        return t('widget.meters.title');
      },
      get description() {
        return t('widget.meters.description');
      },
      defaultSize: { w: 4, h: 6 },
    },
    icon: 'gauge',
    compact: () => 'meters',
    compactExample: () => t('compact.example.meters'),
    view: lazy(
      () => import('../../features/meters/ui/MetersWidget.svelte'),
      (_config, c) => ({ meters: c.meters }),
    ),
  },
];
