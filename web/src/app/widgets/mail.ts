import { t } from '../../shared/i18n';
import { lazy, type WidgetDefinition } from '../widgetKit';

/** The mail widget (#283). */
export const widgets: WidgetDefinition[] = [
  {
    type: {
      kind: 'mail',
      category: 'everyday',
      keywords: ['email', 'e-mail', 'inbox', 'imap', 'posteingang', 'postfach', 'ungelesen', 'unread'],
      get title() {
        return t('widget.mail.title');
      },
      get description() {
        return t('widget.mail.description');
      },
      defaultSize: { w: 4, h: 5 },
    },
    icon: 'note',
    view: lazy(
      () => import('../../features/mail/ui/MailWidget.svelte'),
      (config, c) => ({ config, mail: c.mail }),
    ),
    settings: lazy(
      () => import('../../features/mail/ui/MailWidgetSettings.svelte'),
      (config, c) => ({ config, mail: c.mail }),
    ),
  },
];
