/** One watchlist entry (#284); value is null when the source had no data. */
export interface MarketRow {
  symbol: string;
  value: number | null;
  changePct: number;
  asOf: string;
  history: number[];
  stale: boolean;
  error: string;
}

/** Default watchlist of a new widget. */
export const DEFAULT_SYMBOLS = ['fx:USD', 'stooq:^dax', 'stooq:btcusd'];

/** Widget configuration: the list of symbols. */
export function parseConfig(raw: Record<string, unknown>): { symbols: string[] } {
  const list = Array.isArray(raw['symbols'])
    ? raw['symbols'].filter((s): s is string => typeof s === 'string')
    : [];
  return { symbols: list.length ? list.slice(0, 12) : DEFAULT_SYMBOLS };
}

/** Lines of the settings text area to symbols (empty lines dropped). */
export const symbolsFromText = (text: string): string[] =>
  text
    .split(/[\n,]/)
    .map((s) => s.trim())
    .filter(Boolean);

/** A readable name: "EUR/USD", "AAPL.US", "^DAX". */
export function displayName(symbol: string): string {
  const [src, code = ''] = symbol.split(':');
  return src === 'fx' ? `EUR/${code.toUpperCase()}` : code.toUpperCase();
}

/** Value with sensible decimals for its size. */
export function formatValue(v: number, locale: string): string {
  const digits = Math.abs(v) >= 1000 ? 0 : Math.abs(v) >= 10 ? 2 : 4;
  return v.toLocaleString(locale, {
    minimumFractionDigits: Math.min(digits, 2),
    maximumFractionDigits: digits,
  });
}

/** "+1.23 %" / "−0.50 %". */
export function formatChange(pct: number, locale: string): string {
  const s = Math.abs(pct).toLocaleString(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  return `${pct > 0 ? '+' : pct < 0 ? '−' : '±'}${s} %`;
}

/** Shifts a series to start at zero so a sparkline shows its range, not its level. */
export function relative(values: readonly number[]): { values: number[]; max: number } {
  if (values.length === 0) return { values: [], max: 1 };
  const min = Math.min(...values);
  const max = Math.max(...values);
  return { values: values.map((v) => v - min), max: max - min || 1 };
}
