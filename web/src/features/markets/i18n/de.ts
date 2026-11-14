import type { en } from './en';

/** German UI texts of the markets widget; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.markets.title': 'Märkte',
  'widget.markets.description': 'Währungen, Aktien, Indizes und Krypto mit Tagesveränderung und Verlauf.',
  'markets.symbols': 'Symbole (eines pro Zeile)',
  'markets.symbols_hint':
    'fx:USD für den Euro in einer Währung (EZB); stooq:aapl.us, stooq:^dax oder stooq:btcusd für Aktien, Indizes und Krypto (Stooq). Höchstens 12.',
  'markets.stale': 'veraltet',
  'markets.no_data': 'Keine Daten für dieses Symbol.',
  'markets.unavailable': 'Die Quelle antwortet nicht.',
  'markets.closes_note': 'Tagesschlusskurse, keine Live-Kurse.',
  'markets.trend': '{name}, letzte 30 Tage',
};
