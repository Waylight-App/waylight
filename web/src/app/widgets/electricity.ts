import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The electricity price widget (#288). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'electricity',
      category: 'everyday',
      keywords: ['power', 'energy', 'strom', 'strompreis', 'tibber', 'awattar', 'dynamic tariff', 'börse'],
      get title() {
        return t('widget.electricity.title');
      },
      get description() {
        return t('widget.electricity.description');
      },
      defaultSize: { w: 4, h: 4 },
    },
    icon: 'power',
    view: lazy(
      () => import('../../features/electricity/ui/ElectricityWidget.svelte'),
      (config, c) => ({ config, electricity: c.electricity }),
    ),
    settings: lazy(
      () => import('../../features/electricity/ui/ElectricityWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
