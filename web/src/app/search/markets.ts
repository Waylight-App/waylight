import { convertCurrency, parseCurrencyQuery, type Rates } from '../../features/search/domain/calc/currency';
import { ApiError } from '../../shared/api-client/http';
import { intlLocale, t } from '../../shared/i18n';
import { defineSearch } from '../searchKit';

/**
 * Currency conversion (#482): "50 usd in eur". The rates are loaded once
 * when the search opens (entities() runs then), so the answer itself is
 * immediate like the calculator's.
 */
export const search = defineSearch({
  module: 'markets',
  intents: [],
  source: (c) => {
    let rates: Rates | null = null;
    let problem: 'off' | 'unavailable' | null = null;
    const known = () => new Set(Object.keys(rates?.rates ?? {}));
    const money = (n: number, code: string) =>
      new Intl.NumberFormat(intlLocale(), {
        style: 'currency',
        currency: code,
        maximumFractionDigits: 2,
      }).format(n);
    return {
      id: 'markets.currency',
      async entities() {
        try {
          rates = await c.markets.rates();
          problem = null;
        } catch (e) {
          problem = e instanceof ApiError && e.status === 404 ? 'off' : 'unavailable';
        }
        return [];
      },
      direct(query) {
        if (!rates) {
          // Without rates, only a query that clearly asks for currencies is
          // answered, and only to say why there is no result.
          if (!problem || !parseCurrencyQuery(query, new Set(['EUR', 'USD', 'GBP', 'CHF', 'JPY'])))
            return null;
          return {
            kind: 'text',
            intent: 'markets.currency',
            title: t('smart.currency'),
            text: t(problem === 'off' ? 'smart.currency_off' : 'smart.currency_unavailable'),
            at: null,
          };
        }
        const q = parseCurrencyQuery(query, known());
        if (!q) return null;
        const result = convertCurrency(q, rates);
        if (result === null) return null;
        return {
          kind: 'metric',
          intent: 'markets.currency',
          title: `${money(q.amount, q.from)} → ${q.to}`,
          value: money(result, q.to),
          detail: t(rates.stale ? 'smart.currency_stale' : 'smart.currency_rate', { date: rates.date }),
          copy: String(Number(result.toFixed(2))),
          at: null,
        };
      },
    };
  },
});
