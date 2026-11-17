/** English UI texts of waylight.yaml in the transfer settings (#88). */
export const en = {
  'transfer.config_title': 'Settings as YAML',
  'transfer.config_intro':
    'Appearance, search, your service tiles, bookmarks and the dashboard as one readable file, for Git or a second instance. No secrets; ${NAME} is read from the server environment.',
  'transfer.config_download': 'Download waylight.yaml',
  'transfer.config_apply': 'Apply waylight.yaml…',
  'transfer.config_confirm':
    'Each section in the file replaces those settings completely; sections missing from the file stay as they are. Continue?',
  'transfer.config_managed':
    'Managed by CONFIG_FILE and read-only here: {sections}. Change them in the file and restart Waylight.',
  'transfer.foreign_title': 'Import from another dashboard',
  'transfer.foreign_intro':
    'Adds the service tiles, bookmarks and supported widgets of Homepage or Homarr. Nothing that exists is changed, so importing twice does no harm. You see a preview first.',
  'transfer.foreign_format': 'Dashboard',
  'transfer.foreign_homepage': 'Homepage (gethomepage.dev)',
  'transfer.foreign_homarr': 'Homarr 0.x',
  'transfer.foreign_choose': 'Choose files…',
  'transfer.foreign_hint_homepage':
    'Choose services.yaml, bookmarks.yaml and widgets.yaml from the Homepage config folder.',
  'transfer.foreign_hint_homarr':
    'Choose the board file from Homarr’s data/configs folder. Homarr 1.x exports cannot be read.',
  'transfer.foreign_summary':
    'Adds {tiles} tiles, {bookmarks} bookmarks and {widgets} widgets; {skipped} already exist.',
  'transfer.foreign_tab': 'New tab: {tab}.',
  'transfer.foreign_warnings': '{n} things are not imported',
  'transfer.foreign_yaml': 'Show the result as waylight.yaml',
  'transfer.foreign_apply': 'Import',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
