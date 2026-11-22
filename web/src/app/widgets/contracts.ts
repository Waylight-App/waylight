import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The contracts widget (#687). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'contracts',
      category: 'everyday',
      keywords: [
        'contract',
        'cancel',
        'deadline',
        'subscription',
        'household',
        'vertrag',
        'kündigung',
        'frist',
        'abo',
        'haushalt',
      ],
      get title() {
        return t('widget.contracts.title');
      },
      get description() {
        return t('widget.contracts.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'key',
    compact: () => 'contracts',
    compactExample: () => t('compact.example.contracts'),
    view: lazy(
      () => import('../../features/contracts/ui/ContractsWidget.svelte'),
      (config, c) => ({ config, contracts: c.contracts }),
    ),
    settings: lazy(
      () => import('../../features/contracts/ui/ContractsSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
