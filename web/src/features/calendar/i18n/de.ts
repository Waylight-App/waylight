import type { en } from './en';

/** German texts of the calendar accounts; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'calendar.accounts.title': 'Google- und Microsoft-Konten',
  'calendar.accounts.intro':
    'Zeige deine Google- oder Microsoft-365-Kalender und trage dort Termine ein. Du meldest dich mit einer eigenen App-Registrierung an; vorher fragt Waylight nichts bei Google oder Microsoft an.',
  'calendar.accounts.provider': 'Dienst',
  'calendar.accounts.register':
    'Lege bei {service} eine App-Registrierung an (OAuth-Client, Typ Webanwendung) und trage diese Weiterleitungsadresse ein:',
  'calendar.accounts.console': 'Konsole von {service} öffnen',
  'calendar.accounts.client_id': 'Client-ID',
  'calendar.accounts.client_secret': 'Client-Secret',
  'calendar.accounts.sign_in': 'Mit {service} anmelden',
  'calendar.accounts.signing_in': 'Schließe die Anmeldung im neuen Fenster ab …',
  'calendar.accounts.sign_in_failed':
    'Die Anmeldung hat nicht geklappt. Prüfe Client-ID, Client-Secret und die eingetragene Weiterleitungsadresse.',
  'calendar.accounts.connected': 'Verbunden. Die Kalender erscheinen im Kalender-Widget.',
  'calendar.accounts.popup_blocked':
    'Der Browser hat das Anmeldefenster blockiert. Erlaube Pop-ups für Waylight und versuche es erneut.',
  'calendar.accounts.unavailable':
    'Die Anmeldung bei Google oder Microsoft braucht PUBLIC_URL, eine https://-Adresse, an die die Anbieter den Browser zurückschicken können.',
  'calendar.accounts.disconnect': 'Trennen',
  'calendar.accounts.disconnect_confirm':
    '{name} trennen? Die Kalender verschwinden aus Waylight, und der Zugriff wird widerrufen.',
  'calendar.accounts.read_only': 'schreibgeschützt',
  'calendar.setup_service': 'Nextcloud-Kalender',
  'calendar.setup_server': 'Nextcloud-Adresse',
  'calendar.setup_step_settings':
    'Öffne Nextcloud, klicke auf dein Profilbild und wähle Persönliche Einstellungen.',
  'calendar.setup_step_security': 'Öffne Sicherheit und scrolle zu Geräte & Sitzungen.',
  'calendar.setup_step_create': 'Gib Waylight als App-Name ein und klicke Neues App-Passwort erstellen.',
  'calendar.setup_step_copy':
    'Kopiere den angezeigten Benutzernamen und das App-Passwort und füge sie unten ein.',
  'calendar.setup_username': 'Benutzername',
  'calendar.setup_password': 'App-Passwort',
  'calendar.setup_calendars': 'Kalender: {list}',
};
