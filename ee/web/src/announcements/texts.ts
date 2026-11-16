import type { Catalogue } from '../lib/texts';

/** Texts of the announcements widget (#297). */
export const texts: Catalogue = {
  title: { en: 'Announcements', de: 'Ankündigungen' },
  description: {
    en: 'Notices for everyone, with read marks and expiry.',
    de: 'Hinweise für alle, mit Gelesen-Markierung und Ablauf.',
  },
  write: { en: 'New announcement', de: 'Neue Ankündigung' },
  edit: { en: 'Edit', de: 'Bearbeiten' },
  delete: { en: 'Delete', de: 'Löschen' },
  delete_confirm: { en: 'Delete "{title}"?', de: '„{title}“ löschen?' },
  save: { en: 'Publish', de: 'Veröffentlichen' },
  cancel: { en: 'Cancel', de: 'Abbrechen' },
  field_title: { en: 'Title', de: 'Titel' },
  field_body: { en: 'Text (Markdown)', de: 'Text (Markdown)' },
  field_pinned: { en: 'Pin to the top', de: 'Oben anheften' },
  field_expires: { en: 'Show until (optional)', de: 'Anzeigen bis (optional)' },
  field_notify: {
    en: 'Also send to the notification channels',
    de: 'Auch an die Benachrichtigungskanäle senden',
  },
  mark_read: { en: 'Mark as read', de: 'Als gelesen markieren' },
  unread: { en: '{n} unread', de: '{n} ungelesen' },
  pinned: { en: 'Pinned', de: 'Angeheftet' },
  by: { en: '{name}, {date}', de: '{name}, {date}' },
  until: { en: 'until {date}', de: 'bis {date}' },
  empty: { en: 'No announcements.', de: 'Keine Ankündigungen.' },
  archive: { en: 'Archive', de: 'Archiv' },
  current: { en: 'Current', de: 'Aktuell' },
  archive_empty: {
    en: 'No expired announcements.',
    de: 'Keine abgelaufenen Ankündigungen.',
  },
  read_only: {
    en: 'The licence has expired: announcements stay readable, new ones cannot be written.',
    de: 'Die Lizenz ist abgelaufen: Ankündigungen bleiben lesbar, neue lassen sich nicht schreiben.',
  },
};
