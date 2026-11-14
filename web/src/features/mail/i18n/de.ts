import type { en } from './en';

/** German UI texts of the mail widget; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.mail.title': 'E-Mail',
  'widget.mail.description': 'Anzahl und neueste ungelesene Nachrichten deiner Postfächer.',
  'settings.section.mail': 'E-Mail',
  'mail.unread': '{n} ungelesen',
  'mail.none': 'Keine ungelesenen Mails.',
  'mail.open_webmail': 'Webmail öffnen',
  'mail.failed': 'Nicht lesbar: {names}',
  'mail.account': 'Konto',
  'mail.all_accounts': 'Alle Konten',
  'mail.intro':
    'Nur lesend: Nichts wird als gelesen markiert, verschoben oder gelöscht, und gelesen werden nur Absender, Betreff und Datum, nie der Text. Nutze ein App-Passwort.',
  'mail.guests':
    'Gäste, die einen Tab mit einem E-Mail-Widget sehen, sehen darin Absender und Betreffzeilen. Lege das Widget auf einen eigenen Tab, wenn dir das wichtig ist.',
  'mail.accounts': 'Konten',
  'mail.no_accounts': 'Noch kein Konto.',
  'mail.add': 'Konto hinzufügen',
  'mail.name': 'Name',
  'mail.host': 'IMAP-Server',
  'mail.port': 'Port',
  'mail.username': 'Benutzername',
  'mail.password': 'App-Passwort',
  'mail.folder': 'Ordner',
  'mail.webmail': 'Webmail-Adresse (optional)',
  'mail.save': 'Testen und speichern',
  'mail.saved': 'Verbunden und gespeichert.',
  'mail.remove': '{name} entfernen',
  'mail.fill_in': 'Gib Name, IMAP-Server und Benutzernamen ein.',
  'mail.password_needed': 'Gib das App-Passwort ein.',
};
