import { describe, expect, it } from 'vitest';
import { type Contract, monthlyTotal, parseCents, parseConfig, validateContract } from './contracts';

describe('contracts domain', () => {
  it('parses amounts in German and English notation', () => {
    expect(parseCents('19,99')).toBe(1999);
    expect(parseCents('19.9')).toBe(1990);
    expect(parseCents('1.234,50')).toBe(123450);
    expect(parseCents('1,234.50')).toBe(123450);
    expect(parseCents('12')).toBe(1200);
    expect(parseCents('')).toBe(0);
    expect(parseCents('abc')).toBeNull();
    expect(parseCents('-5')).toBeNull();
  });

  it('adds up the running contracts per month', () => {
    const c = (monthlyCents: number, open: boolean) => ({ monthlyCents, open }) as Contract;
    expect(monthlyTotal([c(1999, true), c(1000, true), c(5000, false)])).toBe(2999);
  });

  it('reads the currency setting', () => {
    expect(parseConfig({})).toEqual({ currency: 'EUR' });
    expect(parseConfig({ currency: 'CHF' })).toEqual({ currency: 'CHF' });
    expect(parseConfig({ currency: 'francs' })).toEqual({ currency: 'EUR' });
  });

  it('validates the dialog', () => {
    const ok = {
      name: 'Gym',
      provider: '',
      cost: '34,90',
      period: 'monthly' as const,
      termEnd: '2027-03-31',
      noticeDays: 90,
      renewMonths: 12,
      remindDays: 14,
      note: '',
    };
    expect(validateContract(ok)).toBeNull();
    expect(validateContract({ ...ok, name: '' })).toBe('contracts.error.name');
    expect(validateContract({ ...ok, cost: 'free' })).toBe('contracts.error.cost');
    expect(validateContract({ ...ok, termEnd: '' })).toBe('contracts.error.term_end');
    expect(validateContract({ ...ok, noticeDays: 800 })).toBe('contracts.error.numbers');
  });
});
