/** English UI texts of compact widget views (#595). */
export const en = {
  'compact.loading': 'Loading…',
  'compact.unavailable': 'Not available',
  'compact.level_attention': 'Needs attention',
  'compact.level_problem': 'Problem',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
