import type { ToolEnv } from './detect';

/** A generated password and the settings it was made with (#477). */
export interface PasswordResult {
  kind: 'password';
  value: string;
  length: number;
  symbols: boolean;
  digitsOnly: boolean;
}

const LOWER = 'abcdefghijkmnopqrstuvwxyz'; // no l: easily confused with 1 and I
const UPPER = 'ABCDEFGHJKLMNPQRSTUVWXYZ'; // no I and O
const DIGITS = '23456789'; // no 0 and 1
const SYMBOLS = '!#$%&*+-=?@_~';

export const DEFAULT_LENGTH = 20;
const MIN_LENGTH = 4;
const MAX_LENGTH = 128;

const NO_SYMBOLS = /\b(?:ohne|keine?|no|without)\s+(?:sonderzeichen|symbole?|symbols?|special)/i;
const DIGITS_ONLY = /\b(?:pin|nur\s+zahlen|nur\s+ziffern|digits?\s+only|only\s+digits|numeric)\b/i;

/**
 * Generates a password for "password", "password 24" or
 * "pw 32 ohne sonderzeichen". Without a length it uses 20 characters; a
 * length outside 4–128 is not a request for a password, so the query falls
 * through to the normal search.
 */
export function password(query: string, env: ToolEnv): PasswordResult | null {
  const lengths = query.match(/\b\d+\b/g) ?? [];
  if (lengths.length > 1) return null;
  const length = lengths[0] === undefined ? DEFAULT_LENGTH : Number(lengths[0]);
  if (length < MIN_LENGTH || length > MAX_LENGTH) return null;
  const digitsOnly = DIGITS_ONLY.test(query);
  const symbols = !digitsOnly && !NO_SYMBOLS.test(query);
  const sets = digitsOnly ? ['0123456789'] : [LOWER, UPPER, DIGITS, ...(symbols ? [SYMBOLS] : [])];
  return { kind: 'password', value: generate(length, sets, env), length, symbols, digitsOnly };
}

/**
 * A random string of the given length with at least one character of every
 * set (when it is long enough), each position drawn without bias.
 */
export function generate(length: number, sets: string[], env: ToolEnv): string {
  const all = sets.join('');
  const chars = Array.from({ length }, () => pick(all, env));
  if (length >= sets.length) {
    // Put one character of each set at distinct random positions.
    const positions = shuffle(
      Array.from({ length }, (_, i) => i),
      env,
    );
    sets.forEach((set, i) => (chars[positions[i] as number] = pick(set, env)));
  }
  return chars.join('');
}

/** A uniformly random index below n, by rejection sampling (no modulo bias). */
function randomIndex(n: number, env: ToolEnv): number {
  const limit = 256 - (256 % n);
  for (;;) {
    for (const b of env.random(16)) if (b < limit) return b % n;
  }
}

function pick(set: string, env: ToolEnv): string {
  return set[randomIndex(set.length, env)] as string;
}

function shuffle<T>(items: T[], env: ToolEnv): T[] {
  for (let i = items.length - 1; i > 0; i--) {
    const j = randomIndex(i + 1, env);
    [items[i], items[j]] = [items[j] as T, items[i] as T];
  }
  return items;
}
