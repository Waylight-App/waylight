import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

// The phone list (#300): an organisation's directory, under "everyday".
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'phonelist',
      category: 'everyday',
      keywords: ['phone', 'telefon', 'contacts', 'kontakte', 'directory', 'verzeichnis'],
      get title() {
        return t('widget.phonelist.title');
      },
      get description() {
        return t('widget.phonelist.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'phone',
    view: lazy(
      () => import('../../features/phonelist/ui/PhonelistWidget.svelte'),
      (_config, c) => ({ phonelist: c.phonelist, admin: c.auth.admin }),
    ),
  },
];
