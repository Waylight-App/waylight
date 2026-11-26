/** English UI texts added to the system feature (#123, #193). */
export const en = {} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
