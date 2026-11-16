import { describe, expect, it } from 'vitest';
import { convertCurrency, parseCurrencyQuery, type Rates } from './currency';

const known = new Set(['EUR', 'USD', 'GBP', 'CHF', 'JPY']);
const rates: Rates = {
  base: 'EUR',
  date: '2026-11-13',
  rates: { EUR: 1, USD: 1.08, GBP: 0.86, CHF: 0.94 },
  stale: false,
};

describe('currency conversion in the search (#482)', () => {
  it('parses codes, symbols and names in German and English', () => {
    expect(parseCurrencyQuery('50 usd in eur', known)).toEqual({ amount: 50, from: 'USD', to: 'EUR' });
    expect(parseCurrencyQuery('50usd to eur', known)).toEqual({ amount: 50, from: 'USD', to: 'EUR' });
    expect(parseCurrencyQuery('€20 in $', known)).toEqual({ amount: 20, from: 'EUR', to: 'USD' });
    expect(parseCurrencyQuery('12,50 € → usd', known)).toEqual({ amount: 12.5, from: 'EUR', to: 'USD' });
    expect(parseCurrencyQuery('100 franken nach euro', known)).toEqual({
      amount: 100,
      from: 'CHF',
      to: 'EUR',
    });
    expect(parseCurrencyQuery('3 Pfund in Euro', known)).toEqual({ amount: 3, from: 'GBP', to: 'EUR' });
  });

  it('stays silent for other input', () => {
    for (const q of ['5 km in mi', '50 usd', 'usd in eur', '50 xyz in eur', '50 eur in eur', '2+3']) {
      expect(parseCurrencyQuery(q, known)).toBeNull();
    }
  });

  it('converts through the euro', () => {
    expect(convertCurrency({ amount: 108, from: 'USD', to: 'EUR' }, rates)).toBeCloseTo(100);
    expect(convertCurrency({ amount: 100, from: 'CHF', to: 'GBP' }, rates)).toBeCloseTo((100 / 0.94) * 0.86);
    expect(convertCurrency({ amount: 1, from: 'JPY', to: 'EUR' }, rates)).toBeNull();
  });
});
