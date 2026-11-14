import { describe, expect, it } from 'vitest';
import { displayName, formatChange, parseConfig, relative, symbolsFromText } from './markets';

describe('markets (#284)', () => {
  it('reads and edits the watchlist', () => {
    expect(parseConfig({}).symbols).toHaveLength(3);
    expect(parseConfig({ symbols: ['fx:CHF', 3] }).symbols).toEqual(['fx:CHF']);
    expect(symbolsFromText('fx:USD\n\n stooq:aapl.us , stooq:^dax')).toEqual([
      'fx:USD',
      'stooq:aapl.us',
      'stooq:^dax',
    ]);
  });
  it('formats names, changes and sparklines', () => {
    expect(displayName('fx:usd')).toBe('EUR/USD');
    expect(displayName('stooq:^dax')).toBe('^DAX');
    expect(formatChange(1.234, 'en')).toBe('+1.23 %');
    expect(formatChange(-0.5, 'en')).toBe('−0.50 %');
    expect(relative([100, 110, 105])).toEqual({ values: [0, 10, 5], max: 10 });
    expect(relative([5, 5])).toEqual({ values: [0, 0], max: 1 });
  });
});
