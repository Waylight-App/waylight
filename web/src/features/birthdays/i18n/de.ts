import type { en } from './en';

/** German UI texts of the birthdays widget; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.birthdays.title': 'Geburtstage',
  'widget.birthdays.description':
    'Die nächsten Geburtstage mit Alter – eingetragen, aus einer vCard-Datei oder deinem Adressbuch.',
  'settings.section.birthdays': 'Geburtstage',
  'birthdays.empty': 'Noch keine Geburtstage.',
  'birthdays.new': 'Geburtstag eintragen',
  'birthdays.name': 'Name',
  'birthdays.date': 'Geburtsdatum',
  'birthdays.year_known': 'Ich kenne das Jahr',
  'birthdays.today': 'heute',
  'birthdays.tomorrow': 'morgen',
  'birthdays.in_days': 'in {n} Tagen',
  'birthdays.turns': 'wird {age}',
  'birthdays.remove': 'Geburtstag von {name} löschen',
  'birthdays.from_book': 'Aus dem Adressbuch',
  'birthdays.invalid_name': 'Gib einen Namen ein (höchstens 100 Zeichen).',
  'birthdays.invalid_date': 'Gib ein Datum ein.',
  'birthdays.limit': 'Anzahl angezeigter Geburtstage',
  'birthdays.import_title': 'vCard-Datei importieren',
  'birthdays.import_hint':
    'Eine .vcf-Datei vom Handy, aus Google Kontakte, Outlook oder Thunderbird. Kontakte ohne Geburtstag und Personen, die schon in der Liste stehen, werden übersprungen.',
  'birthdays.import_button': 'Datei wählen …',
  'birthdays.imported': '{added} hinzugefügt, {duplicates} übersprungen.',
  'birthdays.book_title': 'CardDAV-Adressbuch',
  'birthdays.book_hint':
    'Wird beim Speichern und danach alle sechs Stunden gelesen. Einträge daraus änderst du im Adressbuch, nicht hier. Nutze ein App-Passwort.',
  'birthdays.book_url': 'Adressbuch-URL',
  'birthdays.username': 'Benutzername',
  'birthdays.password': 'App-Passwort',
  'birthdays.password_stored': 'gespeichert – leer lassen, um es zu behalten',
  'birthdays.remind_title': 'Erinnerung',
  'birthdays.remind_on': 'Push am Tag des Geburtstags',
  'birthdays.remind_at': 'Um',
  'birthdays.saved': 'Gespeichert.',
};
