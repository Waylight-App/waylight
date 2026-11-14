/** English UI texts of the electricity price widget (#288). */
export const en = {
  'widget.electricity.title': 'Electricity prices',
  'widget.electricity.description':
    'Day-ahead price per hour for today and tomorrow, with the cheapest hours.',
  'electricity.zone': 'Bidding zone',
  'electricity.zone_de': 'Germany and Luxembourg',
  'electricity.zone_at': 'Austria',
  'electricity.now': 'Now',
  'electricity.today': 'Today',
  'electricity.tomorrow': 'Tomorrow',
  'electricity.cheapest': 'Cheapest hours',
  'electricity.stale': 'The source does not answer; these are the last known prices.',
  'electricity.market_note': 'Market price without taxes and fees.',
  'electricity.hour': '{time}: {price}',
  'electricity.tomorrow_later': 'Tomorrow’s prices follow around 1 pm.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
