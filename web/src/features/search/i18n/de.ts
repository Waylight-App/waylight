import type { en } from './en';

/** German UI texts of the search tools (#477). */
export const de: Record<keyof typeof en, string> = {
  'tools.new': 'Neu',
  'tools.password': 'Passwort · {length} Zeichen',
  'tools.password_all': 'Buchstaben, Ziffern und Sonderzeichen',
  'tools.password_no_symbols': 'Buchstaben und Ziffern',
  'tools.password_digits': 'Nur Ziffern',
  'tools.password_note': 'im Browser erzeugt und nirgends gespeichert',
  'tools.uuid': 'UUID v{version}',
  'tools.colour': 'Farbe',
  'tools.unix_to_date': 'Unix-Zeit {seconds}',
  'tools.date_to_unix': 'Unix-Zeit',
  'tools.local': 'Ortszeit',
  'tools.network': 'Netz {network}',
  'tools.mask': 'Maske',
  'tools.broadcast': 'Broadcast',
  'tools.first_host': 'Erster Host',
  'tools.last_host': 'Letzter Host',
  'tools.hosts': 'Nutzbare Hosts',
  'tools.addresses': 'Adressen',
};
