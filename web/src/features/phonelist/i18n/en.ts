/** English UI texts of the phone list feature (#300). */
export const en = {
  'widget.phonelist.title': 'Phone list',
  'widget.phonelist.description':
    'Who to call: names, roles, phone numbers and e-mail addresses, searchable.',
  'settings.section.phonelist': 'Phone list',
  'phonelist.from_book': 'From the address book; change it there',
  'phonelist.book_title': 'Address book',
  'phonelist.book_hint':
    'Read people from a CardDAV address book, for example a shared one in Nextcloud. They appear next to the entries typed in here, are read every hour and can only be changed in the address book.',
  'phonelist.book_url': 'CardDAV address book',
  'phonelist.book_username': 'User name',
  'phonelist.book_password': 'App password',
  'phonelist.book_password_stored': 'Stored; leave empty to keep it',
  'phonelist.book_last_sync': 'Last read: {time}',
  'phonelist.book_error': 'The last attempt failed: {error}',
  'phonelist.search': 'Search name, role or team',
  'phonelist.add': 'Add a person',
  'phonelist.edit': 'Edit {name}',
  'phonelist.name': 'Name',
  'phonelist.role': 'Role',
  'phonelist.team': 'Team',
  'phonelist.phone': 'Phone',
  'phonelist.email': 'E-mail',
  'phonelist.delete': 'Delete',
  'phonelist.confirm_delete': 'Remove {name} from the phone list?',
  'phonelist.empty': 'The phone list is empty.',
  'phonelist.empty_admin': 'The phone list is empty. Add people with ＋.',
  'phonelist.none_match': 'Nobody matches.',
  'phonelist.no_contact': 'No phone number or e-mail',
  'phonelist.copy': 'Copy {phone}',
  'phonelist.copied': 'Copied',
  'phonelist.group_by_team': 'Group by team',
  'phonelist.other_team': 'Other',
  'phonelist.download_csv': 'Download as CSV',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
