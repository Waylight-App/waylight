/** English UI texts of the shopping lists on Waylight (#471). */
export const en = {
  'shopping.list': 'List',
  'shopping.caldav_list': 'CalDAV task list',
  'shopping.no_list': 'Your shopping list, always at hand.',
  'shopping.new_list_name': 'Name of the new list',
  'shopping.default_list_name': 'Shopping',
  'shopping.create_list': 'Create list',
  'shopping.remove_item': 'Remove “{name}”',
  'shopping.lists_title': 'Lists on Waylight',
  'shopping.lists_hint':
    'These lists need no other service. Share a list with a group or with everyone, and they can add, tick off and remove items too. The widget shows one list; pick it in the widget.',
  'shopping.list_name': 'Name',
  'shopping.open_count': '{count} open',
  'shopping.shared_with_you': 'Shared with you',
  'shopping.shared_with_you_via': 'Shared with you via {group}',
  'shopping.shared_with_everyone': 'Shared with everyone',
  'shopping.save_name': 'Rename',
  'shopping.delete_list': 'Delete',
  'shopping.confirm_delete': 'Delete the list “{name}” and all its items?',
  'shopping.export': 'Export',
  'shopping.export_text': 'as text',
  'shopping.export_ics': 'as .ics',
  'shopping.import_caldav': 'Import from CalDAV',
  'shopping.imported': '{count} open items copied from the CalDAV list.',
  'shopping.caldav_title': 'CalDAV task list',
  'shopping.connect_list': 'Connect a list',
  'shopping.setup_list_url': 'Address of the task list',
  'shopping.setup_list_hint': 'In Nextcloud Tasks: open the ⋯ menu of the list and choose Copy private link.',
  'shopping.setup_username': 'User name',
  'shopping.setup_password': 'App password',
  'shopping.setup_step_security':
    'Open Nextcloud, click your profile picture, choose Personal settings and then Security.',
  'shopping.setup_step_create':
    'Under Devices & sessions enter Waylight as app name and press Create new app password.',
  'shopping.setup_step_copy': 'Copy the user name and the app password shown and paste them below.',
  'shopping.setup_found': '{n} open items',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
