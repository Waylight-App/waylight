/** English UI texts of the in-app help (#678). */
export const en = {
  'help.title': 'About {name}',
  'help.shows': 'What it shows',
  'help.set_up': 'How to set it up',
  'help.docs': 'Read more in the documentation',
  'help.loading': 'Loading…',
  'help.unavailable': 'The help could not be loaded. The documentation has the details.',
  'help.docs_page': 'Documentation',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
