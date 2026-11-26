/** English UI texts of the Tonight card. */
export const en = {
  'tonight.label': 'Time and weather',
  'tonight.choose_a_location': 'Choose a location in the widget settings',
  'tonight.weather_unavailable': 'Weather unavailable',
  'tonight.loading': 'Loading weather',
  'tonight.more_warnings': '+{n} more',
  'tonight.warning': 'Weather warning',
  'tonight.title_in': 'Tonight in {place}',
  'tonight.title': 'Tonight',
  'tonight.feels': 'feels {t}',
  'tonight.forecast': 'Forecast',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
