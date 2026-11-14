import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The markets widget (#284). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'markets',
      category: 'information',
      keywords: [
        'stocks',
        'shares',
        'aktien',
        'börse',
        'kurse',
        'currency',
        'währung',
        'crypto',
        'bitcoin',
        'dax',
      ],
      get title() {
        return t('widget.markets.title');
      },
      get description() {
        return t('widget.markets.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'activity',
    view: lazy(
      () => import('../../features/markets/ui/MarketsWidget.svelte'),
      (config, c) => ({ config, markets: c.markets }),
    ),
    settings: lazy(
      () => import('../../features/markets/ui/MarketsWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
