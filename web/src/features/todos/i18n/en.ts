/** English UI texts of connected to-do lists (#169). */
export const en = {
  'settings.section.todos': 'To-do lists',
  'todos.changed_elsewhere': 'This to-do was changed in the other app meanwhile. Its version is shown now.',
  'todos.sources.intro':
    'Show a task list from Nextcloud Tasks, Synology, Radicale, Baïkal, Fastmail or another CalDAV server in a to-do widget. Changes go both ways; the widget checks for changes every minute.',
  'todos.sources.connected': 'Connected lists',
  'todos.sources.none': 'No list connected yet.',
  'todos.sources.add': 'Connect a list',
  'todos.sources.name': 'Name in trisse-home',
  'todos.sources.url': 'Server or list address',
  'todos.sources.url_hint':
    'The DAV address of the server is enough, e.g. https://cloud.example.com/remote.php/dav – then choose the list below.',
  'todos.sources.username': 'User name',
  'todos.sources.password': 'App password',
  'todos.sources.find': 'Find lists',
  'todos.sources.found': 'Task list',
  'todos.sources.none_found': 'No task lists were found at this address.',
  'todos.sources.connect': 'Connect',
  'todos.sources.remove': 'Disconnect {name}',
  'todos.sources.removed': 'Disconnected. Nothing was deleted on the server.',
  'todos.sources.saved': 'Connected and saved.',
  'todos.sources.limits':
    'Repeating to-dos, reminders and manual ordering stay with the built-in list. Due dates are stored without a time.',
  'todos.sources.url_needed': 'Enter the address of the server or of the list (https://…).',
  'todos.sources.login_needed': 'User name and app password are required.',
  'todos.sources.name_needed': 'Give the list a name.',
  'todos.sources.list': 'List',
  'todos.sources.local': 'Built-in list',
  'todos.sources.missing': 'Disconnected list',
  'todos.sources.widget_hint': 'Connect lists under Settings → To-do lists.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
