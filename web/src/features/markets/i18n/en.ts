/** English UI texts of the markets widget (#284). */
export const en = {
  'widget.markets.title': 'Markets',
  'widget.markets.description': 'Currencies, shares, indices and crypto with daily change and sparkline.',
  'markets.symbols': 'Symbols (one per line)',
  'markets.symbols_hint':
    'fx:USD for the euro in a currency (ECB); stooq:aapl.us, stooq:^dax or stooq:btcusd for shares, indices and crypto (Stooq). At most 12.',
  'markets.stale': 'outdated',
  'markets.no_data': 'No data for this symbol.',
  'markets.unavailable': 'The source does not answer.',
  'markets.closes_note': 'Daily closes, no live prices.',
  'markets.trend': '{name}, last 30 days',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
