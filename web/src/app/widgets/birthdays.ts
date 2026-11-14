import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The birthdays widget (#287). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'birthdays',
      category: 'everyday',
      keywords: ['birthday', 'geburtstag', 'anniversary', 'contacts', 'kontakte', 'vcard', 'carddav'],
      get title() {
        return t('widget.birthdays.title');
      },
      get description() {
        return t('widget.birthdays.description');
      },
      defaultSize: { w: 3, h: 4 },
    },
    icon: 'calendar',
    view: lazy(
      () => import('../../features/birthdays/ui/BirthdaysWidget.svelte'),
      (config, c) => ({ config, birthdays: c.birthdays }),
    ),
    settings: lazy(
      () => import('../../features/birthdays/ui/BirthdaysWidgetSettings.svelte'),
      (config) => ({ config }),
    ),
  },
];
