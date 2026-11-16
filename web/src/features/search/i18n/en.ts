/** English UI texts of the search tools (#477). */
export const en = {
  'tools.new': 'New',
  'tools.password': 'Password · {length} characters',
  'tools.password_all': 'Letters, digits and symbols',
  'tools.password_no_symbols': 'Letters and digits',
  'tools.password_digits': 'Digits only',
  'tools.password_note': 'made in your browser and never saved',
  'tools.uuid': 'UUID v{version}',
  'tools.colour': 'Colour',
  'tools.unix_to_date': 'Unix time {seconds}',
  'tools.date_to_unix': 'Unix time',
  'tools.local': 'local',
  'tools.network': 'Network {network}',
  'tools.mask': 'Mask',
  'tools.broadcast': 'Broadcast',
  'tools.first_host': 'First host',
  'tools.last_host': 'Last host',
  'tools.hosts': 'Usable hosts',
  'tools.addresses': 'Addresses',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
