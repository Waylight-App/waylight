import type { en } from './en';

/** German UI texts of the phone list feature (#300). */
export const de: Record<keyof typeof en, string> = {
  'widget.phonelist.title': 'Telefonliste',
  'widget.phonelist.description':
    'Wen du anrufen kannst: Namen, Rollen, Telefonnummern und E-Mail-Adressen, durchsuchbar.',
  'settings.section.phonelist': 'Telefonliste',
  'phonelist.from_book': 'Aus dem Adressbuch; dort ändern',
  'phonelist.book_title': 'Adressbuch',
  'phonelist.book_hint':
    'Personen aus einem CardDAV-Adressbuch lesen, zum Beispiel einem geteilten in Nextcloud. Sie erscheinen neben den hier eingetragenen, werden stündlich gelesen und lassen sich nur im Adressbuch ändern.',
  'phonelist.book_url': 'CardDAV-Adressbuch',
  'phonelist.book_username': 'Benutzername',
  'phonelist.book_password': 'App-Passwort',
  'phonelist.book_password_stored': 'Gespeichert; leer lassen, um es zu behalten',
  'phonelist.book_last_sync': 'Zuletzt gelesen: {time}',
  'phonelist.book_error': 'Der letzte Versuch ist fehlgeschlagen: {error}',
  'phonelist.search': 'Name, Rolle oder Team suchen',
  'phonelist.add': 'Person hinzufügen',
  'phonelist.edit': '{name} bearbeiten',
  'phonelist.name': 'Name',
  'phonelist.role': 'Rolle',
  'phonelist.team': 'Team',
  'phonelist.phone': 'Telefon',
  'phonelist.email': 'E-Mail',
  'phonelist.delete': 'Löschen',
  'phonelist.confirm_delete': '{name} aus der Telefonliste entfernen?',
  'phonelist.empty': 'Die Telefonliste ist leer.',
  'phonelist.empty_admin': 'Die Telefonliste ist leer. Füge Personen mit ＋ hinzu.',
  'phonelist.none_match': 'Niemand passt.',
  'phonelist.no_contact': 'Keine Telefonnummer oder E-Mail',
};
