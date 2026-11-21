/** English UI texts of the attention feature (#592). */
export const en = {
  'settings.section.attention': 'Needs your attention',
  'attention.title': 'Needs your attention',
  'attention.badge': '{n} things need your attention',
  'attention.badge_one': '1 thing needs your attention',
  'attention.intro': 'Widgets and connections that stopped working or are not set up yet.',
  'attention.failing': 'Not working any more',
  'attention.setup': 'Not set up yet',
  'attention.empty': 'Everything works. Nothing needs your attention.',
  'attention.loading': 'Checking…',
  'attention.since': 'since {value}',
  'attention.check_all': 'Check all again',
  'attention.action.signIn': 'Sign in again',
  'attention.action.check': 'Check again',
  'attention.action.fix': 'Fix',
  'attention.action.choose': 'Set up',
  'attention.action.connect': 'Connect',
  'attention.on_tab': 'On tab {value}',
  'attention.instance': 'Used by the whole instance',
  'settings.section.connection_status': 'Connected services',
  'services.intro': 'The services your widgets connect to, and whether they answer.',
  'services.summary': '{connected} connected · {failing} with a problem · {setup} not set up',
  'services.empty': 'No service reports a connection yet.',
  'services.ok': 'Connected',
  'services.setup': 'Not set up yet',
  'services.used_by': 'Used by {list}',
  'services.unused': 'No widget of yours uses it',
  'services.instance': 'for the whole instance',
  'services.open': 'Settings',
  'services.check_all': 'Check all again',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
