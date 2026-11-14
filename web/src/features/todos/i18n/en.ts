/** English UI texts of connected to-do lists (#169). */
export const en = {
  'settings.section.todos': 'To-do lists',
  'todos.changed_elsewhere': 'This to-do was changed in the other app meanwhile. Its version is shown now.',
  'todos.sources.intro':
    'Show a task list from Todoist, Nextcloud Tasks, Synology, Radicale, Baïkal, Fastmail or another CalDAV server in a to-do widget. Changes go both ways; the widget checks for changes every minute.',
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
  'todos.sources.kind': 'Service',
  'todos.sources.kind_caldav': 'CalDAV (Nextcloud, Synology, Radicale, …)',
  'todos.sources.kind_todoist': 'Todoist',
  'todos.sources.kind_microsoft': 'Microsoft To Do',
  'todos.sources.kind_google': 'Google Tasks',
  'todos.sources.kind_ticktick': 'TickTick',
  'todos.sources.oauth_intro':
    'Sign in with an app you register yourself, so trisse-home talks to {service} without any company in between. Register this redirect address:',
  'todos.sources.oauth_console': 'Open the app registration of {service}',
  'todos.sources.oauth_unavailable':
    'Signing in needs PUBLIC_URL, the https:// address of trisse-home (see the configuration guide).',
  'todos.sources.client_id': 'Client ID',
  'todos.sources.client_secret': 'Client secret',
  'todos.sources.client_id_needed': 'Enter the client ID of your app registration.',
  'todos.sources.sign_in': 'Sign in',
  'todos.sources.signing_in': 'Waiting for the sign-in in the other window…',
  'todos.sources.signed_in': 'Signed in. Choose the list below.',
  'todos.sources.sign_in_failed':
    'The sign-in did not succeed. Check client ID, client secret and the redirect address.',
  'todos.sources.popup_blocked':
    'The browser blocked the sign-in window. Allow pop-ups for trisse-home and try again.',
  'todos.sources.token': 'API token',
  'todos.sources.token_hint':
    'In Todoist: Settings → Integrations → Developer → copy the API token. Then find your projects below.',
  'todos.sources.token_needed': 'Enter the Todoist API token.',
  'todos.sources.project': 'Project',
  'todos.sources.project_needed': 'Find your projects and choose one.',
  'todos.sources.limits_todoist':
    'Todoist is asked for changes at most every 15 seconds. Completing a recurring Todoist task moves it to its next date, as in the Todoist apps.',
  'todos.sources.local': 'Built-in list',
  'todos.sources.missing': 'Disconnected list',
  'todos.sources.widget_hint': 'Connect lists under Settings → To-do lists.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
