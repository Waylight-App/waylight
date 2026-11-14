import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The sports widget (#285). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'sports',
      category: 'information',
      keywords: ['football', 'soccer', 'bundesliga', 'fußball', 'fussball', 'tabelle', 'ergebnis', 'results'],
      get title() {
        return t('widget.sports.title');
      },
      get description() {
        return t('widget.sports.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    icon: 'activity',
    view: lazy(
      () => import('../../features/sports/ui/SportsWidget.svelte'),
      (config, c) => ({ config, sports: c.sports }),
    ),
    settings: lazy(
      () => import('../../features/sports/ui/SportsWidgetSettings.svelte'),
      (config, c) => ({ config, sports: c.sports }),
    ),
  },
];
