/**
 * Currency conversion in the search (#482): "50 usd in eur", "€20 to $",
 * "100 franken in euro". Parsing is pure; the rates come from the server
 * (GET /api/markets/rates, the ECB's daily reference rates, euro based).
 */

/** Symbols and names people type, mapped to ISO codes (lower case). */
const ALIASES: Record<string, string> = {
  '€': 'EUR',
  euro: 'EUR',
  euros: 'EUR',
  $: 'USD',
  us$: 'USD',
  dollar: 'USD',
  dollars: 'USD',
  '£': 'GBP',
  pound: 'GBP',
  pounds: 'GBP',
  pfund: 'GBP',
  '¥': 'JPY',
  yen: 'JPY',
  franken: 'CHF',
  franc: 'CHF',
  francs: 'CHF',
  'fr.': 'CHF',
  krone: 'DKK',
  kronen: 'DKK',
  zloty: 'PLN',
  złoty: 'PLN',
  yuan: 'CNY',
  renminbi: 'CNY',
  rupee: 'INR',
  rupees: 'INR',
  rupie: 'INR',
  rupien: 'INR',
};

/** A currency conversion the user asked for. */
export interface CurrencyQuery {
  amount: number;
  from: string;
  to: string;
}

/** Rates as the server sends them: units per euro on a day. */
export interface Rates {
  base: string;
  date: string;
  rates: Record<string, number>;
  stale: boolean;
}

/** Reads a currency token: an ISO code the rates know, a symbol or a name. */
export function currencyCode(token: string, known: ReadonlySet<string>): string | null {
  const t = token.trim().toLowerCase();
  if (!t) return null;
  const alias = ALIASES[t];
  if (alias && known.has(alias)) return alias;
  const iso = t.toUpperCase();
  return /^[A-Z]{3}$/.test(iso) && known.has(iso) ? iso : null;
}

const NUMBER = String.raw`(\d+(?:[.,]\d+)?)`;
const CURRENCY = String.raw`([a-zà-žł$€£¥.]+\$?|[$€£¥])`;
const CONNECT = String.raw`(?:in|to|nach|zu|->|→|=|als)`;
/** "50 usd in eur", "50usd to eur", "€ 50 in $", "50 € → usd". */
const AMOUNT_FIRST = new RegExp(String.raw`^${NUMBER}\s*${CURRENCY}\s+${CONNECT}\s+${CURRENCY}$`, 'iu');
const SYMBOL_FIRST = new RegExp(String.raw`^([$€£¥])\s*${NUMBER}\s+${CONNECT}\s+${CURRENCY}$`, 'iu');

const parseAmount = (s: string): number => Number(s.replace(',', '.'));

/** Parses "50 usd in eur" against the currencies the rates know. */
export function parseCurrencyQuery(input: string, known: ReadonlySet<string>): CurrencyQuery | null {
  const text = input.trim().replace(/\s+/g, ' ');
  if (text.length > 60) return null;
  let amount: number, from: string | null, to: string | null;
  const a = AMOUNT_FIRST.exec(text);
  const s = a ? null : SYMBOL_FIRST.exec(text);
  if (a) {
    amount = parseAmount(a[1] ?? '');
    from = currencyCode(a[2] ?? '', known);
    to = currencyCode(a[3] ?? '', known);
  } else if (s) {
    amount = parseAmount(s[2] ?? '');
    from = currencyCode(s[1] ?? '', known);
    to = currencyCode(s[3] ?? '', known);
  } else {
    return null;
  }
  if (!from || !to || !Number.isFinite(amount) || from === to) return null;
  return { amount, from, to };
}

/** Converts with euro-based rates; null when a currency is unknown. */
export function convertCurrency(q: CurrencyQuery, r: Rates): number | null {
  const from = r.rates[q.from];
  const to = r.rates[q.to];
  if (!from || !to) return null;
  return (q.amount / from) * to;
}
