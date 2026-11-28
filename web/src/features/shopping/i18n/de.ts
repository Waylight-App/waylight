import type { en } from './en';

/** German UI texts of the shopping lists on Waylight (#471). */
export const de: Record<keyof typeof en, string> = {
  'shopping.list': 'Liste',
  'shopping.caldav_list': 'CalDAV-Aufgabenliste',
  'shopping.no_list': 'Deine Einkaufsliste, immer griffbereit.',
  'shopping.new_list_name': 'Name der neuen Liste',
  'shopping.default_list_name': 'Einkaufen',
  'shopping.create_list': 'Liste anlegen',
  'shopping.remove_item': '„{name}“ entfernen',
  'shopping.ticked': '„{name}“ abgehakt',
  'shopping.removed': '„{name}“ entfernt',
  'shopping.lists_title': 'Listen in Waylight',
  'shopping.lists_hint':
    'Diese Listen brauchen keinen anderen Dienst. Teile eine Liste mit einer Gruppe oder mit allen, dann können auch sie Einträge hinzufügen, abhaken und entfernen. Das Widget zeigt eine Liste; wähle sie im Widget aus.',
  'shopping.list_name': 'Name',
  'shopping.open_count': '{count} offen',
  'shopping.shared_with_you': 'Mit dir geteilt',
  'shopping.shared_with_you_via': 'Mit dir geteilt über {group}',
  'shopping.shared_with_everyone': 'Mit allen geteilt',
  'shopping.save_name': 'Umbenennen',
  'shopping.delete_list': 'Löschen',
  'shopping.confirm_delete': 'Die Liste „{name}“ mit allen Einträgen löschen?',
  'shopping.export': 'Exportieren',
  'shopping.export_text': 'als Text',
  'shopping.export_ics': 'als .ics',
  'shopping.import_caldav': 'Aus CalDAV übernehmen',
  'shopping.imported': '{count} offene Einträge aus der CalDAV-Liste übernommen.',
  'shopping.caldav_title': 'CalDAV-Aufgabenliste',
  'shopping.connect_list': 'Liste verbinden',
  'shopping.setup_list_url': 'Adresse der Aufgabenliste',
  'shopping.setup_list_hint':
    'In Nextcloud Aufgaben: das ⋯-Menü der Liste öffnen und „Privaten Link kopieren“ wählen.',
  'shopping.setup_username': 'Benutzername',
  'shopping.setup_password': 'App-Passwort',
  'shopping.setup_step_security':
    'Öffne Nextcloud, klicke auf dein Profilbild, wähle Persönliche Einstellungen und dann Sicherheit.',
  'shopping.setup_step_create':
    'Gib unter Geräte & Sitzungen Waylight als App-Namen ein und klicke Neues App-Passwort erstellen.',
  'shopping.setup_step_copy':
    'Kopiere den angezeigten Benutzernamen und das App-Passwort und füge sie unten ein.',
  'shopping.setup_found': '{n} offene Einträge',
};
