import { t } from '../../shared/i18n';
import { choicesOf, lazy, type WidgetDefinition } from '../widgetKit';

/** The Immich widget. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'immich',
      category: 'network',
      keywords: ['photos', 'fotos', 'gallery', 'memories', 'erinnerungen'],
      get title() {
        return t('widget.immich.title');
      },
      get description() {
        return t('widget.immich.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'archive',
    view: lazy(
      () => import('../../features/immich/ui/ImmichWidget.svelte'),
      (config, c) => ({
        config,
        immich: c.immich,
        discover: () => c.discovery.find('immich'),
        connections: choicesOf(c),
      }),
    ),
    settings: lazy(
      () => import('../../features/immich/ui/ImmichWidgetSettings.svelte'),
      (config, c) => ({ config, connections: choicesOf(c) }),
    ),
  },
];
