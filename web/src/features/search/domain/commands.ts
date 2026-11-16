import { fold, matchToken, QUALITY_FACTOR, tokenize } from './smart/text';

/**
 * App commands of the search field (#476): "dark mode", "einstellungen
 * wetter", "tab arbeit". Commands are data – id, German and English terms,
 * the role that may run them and the key of the handler the composition
 * root provides – so the help (#479) can list them and the search feature
 * imports no other feature. Matching and ranking are pure functions.
 */

/** Who may run a command. Kiosk displays get none. */
export type CommandRole = 'everyone' | 'member' | 'admin' | 'owner';

export interface CommandSpec {
  id: string;
  /** Words or phrases in German and English; phrases are split into words. */
  terms: { de: string[]; en: string[] };
  role: CommandRole;
  /** Key of the handler that runs the command, e.g. "theme" or "settings". */
  handler: string;
  /** Argument for the handler, e.g. the theme mode or the settings section. */
  arg?: string;
  /**
   * One of these words must be typed (outside ">" mode), so generated
   * commands like "Settings: Weather" do not answer every "wetter".
   */
  requires?: string[];
  /** Words that are not terms are passed on, e.g. "add widget weather". */
  rest?: boolean;
  /** Keyboard shortcut shown on the row. */
  shortcut?: string;
  /** Ranks a general command above generated ones with the same words; default 1. */
  weight?: number;
}

/** What the signed-in account is; decides which commands are offered. */
export interface Viewer {
  guest: boolean;
  kiosk: boolean;
  admin: boolean;
  owner: boolean;
}

/** A command that matched the typed text. */
export interface CommandMatch {
  spec: CommandSpec;
  score: number;
  /** The words that are not terms of the command, for rest commands. */
  rest: string;
}

/** The prefix that limits the results to commands (command-palette convention). */
export const COMMAND_PREFIX = '>';

/** Commands shown without ">" and with it. */
export const MAX_COMMANDS = 3;
export const MAX_COMMANDS_ONLY = 8;

/** Shortest text that offers commands without ">". */
export const MIN_QUERY = 3;

/** Whether the account may run a command; kiosk displays may run none. */
export function permitted(role: CommandRole, viewer: Viewer): boolean {
  if (viewer.kiosk) return false;
  switch (role) {
    case 'everyone':
      return true;
    case 'member':
      return !viewer.guest;
    case 'admin':
      return viewer.admin;
    case 'owner':
      return viewer.owner;
  }
}

/** Splits ">theme" into the command-only flag and the text after the prefix. */
export function parseCommandQuery(query: string): { only: boolean; text: string } {
  const trimmed = query.trimStart();
  if (trimmed.startsWith(COMMAND_PREFIX)) return { only: true, text: trimmed.slice(1).trim() };
  return { only: false, text: query.trim() };
}

const words = (spec: CommandSpec) => [
  ...new Set([...spec.terms.de, ...spec.terms.en].flatMap((term) => tokenize(term))),
];

/**
 * Scores a command for the typed words: every word must match one of the
 * command's words (exact, prefix or a small typo), except the rest of a
 * rest command. Null when the command does not match.
 */
export function scoreCommand(spec: CommandSpec, text: string, only = false): CommandMatch | null {
  const tokens = tokenize(text);
  if (tokens.length === 0) return only ? { spec, score: 0, rest: '' } : null;
  const vocabulary = words(spec);
  let score = 0;
  let matched = 0;
  const rest: string[] = [];
  for (const token of tokens) {
    let best = 0;
    for (const word of vocabulary) {
      const q = matchToken(token, word);
      if (q) best = Math.max(best, QUALITY_FACTOR[q]);
    }
    if (best > 0 && rest.length === 0) {
      score += best;
      matched++;
    } else if (spec.rest && matched > 0) rest.push(token);
    else return null;
  }
  if (!only && spec.requires && !tokens.some((t) => spec.requires?.some((r) => matchToken(t, fold(r))))) {
    return null;
  }
  return { spec, score: score * (spec.weight ?? 1), rest: rest.join(' ') };
}

/**
 * The commands offered for the typed text, best first: commands the viewer
 * may not run are left out, the boost of earlier choices (frecency, #159)
 * is applied, and at most MAX_COMMANDS are returned (more after ">").
 */
export function matchCommands(
  specs: readonly CommandSpec[],
  query: string,
  viewer: Viewer,
  boost: (id: string, score: number) => number = (_, s) => s,
): CommandMatch[] {
  const { only, text } = parseCommandQuery(query);
  if (!only && text.length < MIN_QUERY) return [];
  return specs
    .filter((s) => permitted(s.role, viewer))
    .map((s) => scoreCommand(s, text, only))
    .filter((m): m is CommandMatch => m !== null)
    .map((m) => ({ ...m, score: boost(m.spec.id, m.score) }))
    .sort((a, b) => b.score - a.score || a.spec.id.localeCompare(b.spec.id))
    .slice(0, only ? MAX_COMMANDS_ONLY : MAX_COMMANDS);
}
