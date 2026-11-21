/** English UI texts of the attention line (#593). */
export const en = {
  'attention.more': 'and {count} more',
  'attention.view': 'View',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
