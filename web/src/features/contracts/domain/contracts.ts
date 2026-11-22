import type { Visibility } from '../../../shared/sharing/sharing';

/** How often a contract's cost is due. */
export const PERIODS = ['monthly', 'quarterly', 'yearly'] as const;
export type Period = (typeof PERIODS)[number];

/** A household contract with its next cancellation deadline (#687). */
export interface Contract {
  id: string;
  name: string;
  provider: string;
  costCents: number;
  period: Period;
  monthlyCents: number;
  /** End of term as entered, YYYY-MM-DD. */
  termEnd: string;
  noticeDays: number;
  renewMonths: number;
  remindDays: number;
  note: string;
  /** End of the term the next deadline belongs to. */
  currentEnd: string;
  deadline: string;
  /** False when the contract ends and its last deadline passed. */
  open: boolean;
  soon: boolean;
  visibility: Visibility;
  mine: boolean;
}

/** What the dialog edits; the cost as typed, e.g. "19.99". */
export interface ContractInput {
  name: string;
  provider: string;
  cost: string;
  period: Period;
  termEnd: string;
  noticeDays: number;
  renewMonths: number;
  remindDays: number;
  note: string;
}

/** Widget configuration. */
export interface ContractsConfig {
  /** ISO 4217 code the amounts are shown in. */
  currency: string;
}

export const DEFAULT_CURRENCY = 'EUR';

/** Reads the widget configuration with a safe default. */
export function parseConfig(raw: Record<string, unknown>): ContractsConfig {
  const c = raw['currency'];
  return { currency: typeof c === 'string' && /^[A-Z]{3}$/.test(c) ? c : DEFAULT_CURRENCY };
}

/**
 * Parses a typed amount ("19,99", "1.234,50", "19.99") into cents, or null.
 * The last separator followed by one or two digits is the decimal mark.
 */
export function parseCents(text: string): number | null {
  const s = text.trim().replace(/\s/g, '');
  if (s === '') return 0;
  const m = /^(\d{1,3}(?:[.,']\d{3})*|\d+)(?:[.,](\d{1,2}))?$/.exec(s);
  if (!m) return null;
  const whole = Number((m[1] ?? '0').replace(/[.,']/g, ''));
  const frac = Number((m[2] ?? '').padEnd(2, '0') || '0');
  return whole * 100 + frac;
}

/** Cents as a plain amount for the input field ("19.99"). */
export const centsToInput = (cents: number): string => (cents / 100).toFixed(2);

/** Formats cents in the currency for the locale. */
export function formatMoney(cents: number, currency: string, locale: string): string {
  return new Intl.NumberFormat(locale, { style: 'currency', currency }).format(cents / 100);
}

/** Sum of the monthly costs of the contracts that still run. */
export const monthlyTotal = (list: Contract[]): number =>
  list.filter((c) => c.open).reduce((sum, c) => sum + c.monthlyCents, 0);

/** Message keys of validateContract. */
export type ContractError =
  'contracts.error.name' | 'contracts.error.cost' | 'contracts.error.term_end' | 'contracts.error.numbers';

/** Validates the dialog and returns a message key or null. */
export function validateContract(input: ContractInput): ContractError | null {
  const name = input.name.trim();
  if (!name || [...name].length > 80) return 'contracts.error.name';
  const cents = parseCents(input.cost);
  if (cents === null || cents > 100_000_000) return 'contracts.error.cost';
  if (!/^\d{4}-\d{2}-\d{2}$/.test(input.termEnd)) return 'contracts.error.term_end';
  const within = (n: number, max: number) => Number.isInteger(n) && n >= 0 && n <= max;
  if (!within(input.noticeDays, 730) || !within(input.renewMonths, 60) || !within(input.remindDays, 90))
    return 'contracts.error.numbers';
  return null;
}
