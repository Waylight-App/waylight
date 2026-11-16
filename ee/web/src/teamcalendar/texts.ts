import type { Catalogue } from '../lib/texts';

/** Texts of the team calendar section (#299). */
export const texts: Catalogue = {
  label: { en: 'Team calendar', de: 'Teamkalender' },
  description: {
    en: 'Share one of your calendar subscriptions with a group: its members see the events in their calendar widget, read-only. The address of the calendar stays hidden.',
    de: 'Teile eines deiner Kalender-Abos mit einer Gruppe: Ihre Mitglieder sehen die Termine in ihrem Kalender-Widget, nur lesend. Die Adresse des Kalenders bleibt verborgen.',
  },
  hint: {
    en: 'Add the subscription under Settings → Calendar first.',
    de: 'Lege das Abo zuerst unter Einstellungen → Kalender an.',
  },
  subscriptions: { en: 'Subscriptions (ICS)', de: 'Abos (ICS)' },
  accounts: { en: 'Connected calendar accounts', de: 'Verbundene Kalenderkonten' },
  accounts_hint: {
    en: 'Sharing an account shares the calendars you enabled for it.',
    de: 'Wer ein Konto teilt, teilt die Kalender, die dafür eingeschaltet sind.',
  },
  only_me: { en: 'Only me', de: 'Nur ich' },
  everyone: { en: 'Everyone', de: 'Alle' },
  group: { en: 'Group: {name}', de: 'Gruppe: {name}' },
  saved: { en: 'Saved.', de: 'Gespeichert.' },
  none: {
    en: 'You have no calendar subscriptions yet.',
    de: 'Du hast noch keine Kalender-Abos.',
  },
  read_only: {
    en: 'The licence has expired: shared calendars stay visible, but sharing cannot be changed.',
    de: 'Die Lizenz ist abgelaufen: Geteilte Kalender bleiben sichtbar, die Freigabe lässt sich aber nicht ändern.',
  },
};
