/** English UI texts of the phone list feature (#300). */
export const en = {
  'widget.phonelist.title': 'Phone list',
  'widget.phonelist.description':
    'Who to call: names, roles, phone numbers and e-mail addresses, searchable.',
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
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
