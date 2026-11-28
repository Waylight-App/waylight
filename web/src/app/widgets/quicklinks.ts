import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** Quick links (#859): favourite bookmarks as tiles, opened with 1 to 9. */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'quicklinks',
      category: 'everyday',
      keywords: ['links', 'favorites', 'favourites', 'shortcuts', 'schnellzugriff', 'favoriten', 'lesezeichen'],
      get title() {
        return t('widget.quicklinks.title');
      },
      get description() {
        return t('widget.quicklinks.description');
      },
      defaultSize: { w: 6, h: 3 },
    },
    icon: 'star',
    view: lazy(
      () => import('../../features/bookmarks/ui/QuickLinksWidget.svelte'),
      (config, c) => ({ config, bookmarks: c.bookmarks, targets: c.quickLinks }),
    ),
    settings: lazy(
      () => import('../../features/bookmarks/ui/QuickLinksSettings.svelte'),
      (config, c) => ({ config, bookmarks: c.bookmarks }),
    ),
  },
];
