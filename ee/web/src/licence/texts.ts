import type { Catalogue } from '../lib/texts';

/** Texts of the section "Features for organisations". */
export const texts: Catalogue = {
  intro: {
    en: 'Features for organisations are part of this edition. Each one works once a licence key unlocks it; the others stay listed here.',
    de: 'Funktionen für Organisationen gehören zu dieser Ausgabe. Jede funktioniert, sobald ein Lizenzschlüssel sie freischaltet; die übrigen bleiben hier aufgelistet.',
  },
  on: { en: 'On', de: 'An' },
  read: { en: 'Read only', de: 'Nur lesen' },
  off: { en: 'Off', de: 'Aus' },
  key_title: { en: 'Licence key', de: 'Lizenzschlüssel' },
  status_missing: { en: 'No licence key yet.', de: 'Noch kein Lizenzschlüssel.' },
  status_valid: { en: 'Licensed to {licensee} until {until}.', de: 'Lizenziert für {licensee} bis {until}.' },
  status_expired: {
    en: 'The key for {licensee} expired on {until}. Data stays readable; changes are paused.',
    de: 'Der Schlüssel für {licensee} ist am {until} abgelaufen. Daten bleiben lesbar, Änderungen sind pausiert.',
  },
  status_over_seats: {
    en: '{accounts} active accounts, the key covers {seats}. Data stays readable; changes are paused.',
    de: '{accounts} aktive Konten, der Schlüssel deckt {seats}. Daten bleiben lesbar, Änderungen sind pausiert.',
  },
  status_invalid: { en: 'The stored key is not valid: {problem}', de: 'Der gespeicherte Schlüssel ist ungültig: {problem}' },
  seats: { en: 'Accounts: {accounts} of {seats}', de: 'Konten: {accounts} von {seats}' },
  from_env: {
    en: 'The key is set with LICENCE_KEY on the server and can only be changed there.',
    de: 'Der Schlüssel ist auf dem Server mit LICENCE_KEY gesetzt und lässt sich nur dort ändern.',
  },
  paste: { en: 'Paste a licence key', de: 'Lizenzschlüssel einfügen' },
  install: { en: 'Save key', de: 'Schlüssel speichern' },
  remove: { en: 'Remove key', de: 'Schlüssel entfernen' },
  remove_confirm: {
    en: 'Remove the licence key? Licensed features switch off; their data is kept.',
    de: 'Lizenzschlüssel entfernen? Lizenzierte Funktionen schalten ab, ihre Daten bleiben erhalten.',
  },
  saved: { en: 'Saved.', de: 'Gespeichert.' },
  unknown: {
    en: 'The key also unlocks features this version does not have yet: {list}',
    de: 'Der Schlüssel schaltet auch Funktionen frei, die diese Version noch nicht hat: {list}',
  },
  admins_only: { en: 'Only admins can see and change the licence key.', de: 'Nur Admins können den Lizenzschlüssel sehen und ändern.' },
};
