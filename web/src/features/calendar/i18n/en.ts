/** English texts of the Google and Microsoft calendar accounts (#282). */
export const en = {
  'calendar.accounts.title': 'Google and Microsoft accounts',
  'calendar.accounts.intro':
    'Show your Google or Microsoft 365 calendars and add events to them. You sign in with an app registration of your own; Waylight asks nothing of Google or Microsoft before that.',
  'calendar.accounts.provider': 'Service',
  'calendar.accounts.register':
    'Create an app registration (OAuth client, type web application) at {service} and register this redirect address:',
  'calendar.accounts.console': 'Open the {service} console',
  'calendar.accounts.client_id': 'Client ID',
  'calendar.accounts.client_secret': 'Client secret',
  'calendar.accounts.sign_in': 'Sign in with {service}',
  'calendar.accounts.signing_in': 'Finish signing in in the new window …',
  'calendar.accounts.sign_in_failed':
    'The sign-in did not succeed. Check client ID, client secret and the registered redirect address.',
  'calendar.accounts.connected': 'Connected. The calendars show up in the calendar widget.',
  'calendar.accounts.popup_blocked':
    'The browser blocked the sign-in window. Allow pop-ups for Waylight and try again.',
  'calendar.accounts.unavailable':
    'Signing in to Google or Microsoft needs PUBLIC_URL, an https:// address under which the providers can send the browser back.',
  'calendar.accounts.disconnect': 'Disconnect',
  'calendar.accounts.disconnect_confirm':
    'Disconnect {name}? Its calendars disappear from Waylight and the access is revoked.',
  'calendar.accounts.read_only': 'read-only',
  'calendar.setup_service': 'Nextcloud calendar',
  'calendar.setup_server': 'Nextcloud address',
  'calendar.setup_step_settings': 'Open Nextcloud, click your profile picture and choose Personal settings.',
  'calendar.setup_step_security': 'Open Security and scroll down to Devices & sessions.',
  'calendar.setup_step_create': 'Enter Waylight as app name and press Create new app password.',
  'calendar.setup_step_copy': 'Copy the user name and the app password shown and paste them below.',
  'calendar.setup_username': 'User name',
  'calendar.setup_password': 'App password',
  'calendar.setup_calendars': 'Calendars: {list}',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
