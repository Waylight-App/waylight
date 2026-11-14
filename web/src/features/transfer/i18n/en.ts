/** English UI texts of trisse-home.yaml in the transfer settings (#88). */
export const en = {
  'transfer.config_title': 'Settings as YAML',
  'transfer.config_intro':
    'Appearance, search, your service tiles, bookmarks and the dashboard as one readable file, for Git or a second instance. No secrets; ${NAME} is read from the server environment.',
  'transfer.config_download': 'Download trisse-home.yaml',
  'transfer.config_apply': 'Apply trisse-home.yaml…',
  'transfer.config_confirm':
    'Each section in the file replaces those settings completely; sections missing from the file stay as they are. Continue?',
  'transfer.config_managed':
    'Managed by CONFIG_FILE and read-only here: {sections}. Change them in the file and restart trisse-home.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
