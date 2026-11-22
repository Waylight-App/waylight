/** English UI texts of the birthdays widget (#287). */
export const en = {
  'widget.birthdays.title': 'Birthdays',
  'widget.birthdays.description':
    'The next birthdays, with age – typed in, from a vCard file or your address book.',
  'settings.section.birthdays': 'Birthdays',
  'birthdays.empty': 'No birthdays yet.',
  'birthdays.new': 'Add birthday',
  'birthdays.name': 'Name',
  'birthdays.date': 'Date of birth',
  'birthdays.year_known': 'I know the year',
  'birthdays.today': 'today',
  'birthdays.tomorrow': 'tomorrow',
  'birthdays.in_days': 'in {n} days',
  'birthdays.turns': 'turns {age}',
  'birthdays.remove': 'Delete birthday of {name}',
  'birthdays.from_book': 'From the address book',
  'birthdays.invalid_name': 'Enter a name (at most 100 characters).',
  'birthdays.invalid_date': 'Enter a date.',
  'birthdays.limit': 'Number of birthdays shown',
  'birthdays.import_title': 'Import vCard file',
  'birthdays.import_hint':
    'A .vcf file from your phone, Google Contacts, Outlook or Thunderbird. Contacts without a birthday and people already in the list are skipped.',
  'birthdays.import_button': 'Choose file …',
  'birthdays.imported': '{added} added, {duplicates} skipped.',
  'birthdays.book_title': 'CardDAV address book',
  'birthdays.book_hint':
    'Read when you save and then every six hours. Its entries are changed in the address book, not here. Use an app password.',
  'birthdays.book_url': 'Address book server address',
  'birthdays.username': 'User name',
  'birthdays.password': 'App password',
  'birthdays.password_stored': 'stored – leave empty to keep it',
  'birthdays.remind_title': 'Reminder',
  'birthdays.remind_on': 'Push on the day of a birthday',
  'birthdays.remind_at': 'At',
  'birthdays.saved': 'Saved.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
