import { fold } from './smart/text';

/**
 * The help of the search field (#479): "?" or "hilfe" lists what the search
 * can do, built from the same data that drives it (intents, commands,
 * capture prefixes, the calculator and time answers). The composition root
 * turns that data into entries in the current language; this file only
 * reads the query, filters and groups. Pure functions only.
 */

/** Groups in display order. */
export const HELP_GROUPS = ['day', 'home', 'time', 'tools', 'capture', 'commands', 'web'] as const;
export type HelpGroup = (typeof HELP_GROUPS)[number];

/** One thing the search can do, already in the current language. */
export interface HelpEntry {
  id: string;
  group: HelpGroup;
  /** What it does, in plain words. */
  text: string;
  /** Queries that work; choosing one puts it into the search field. */
  examples: string[];
}

export interface HelpSection {
  group: HelpGroup;
  entries: HelpEntry[];
}

/** Words that open the help, followed by an optional filter. */
const HELP = /^(?:\?|hilfe\b|help\b)\s*(.*)$/;

/**
 * Whether the query asks for help and what to filter by: "?" → "",
 * "? wetter" → "wetter", "hilfe müll" → "mull". Null for anything else,
 * so "helpdesk" or "?foo=bar" in a URL stay normal searches.
 */
export function parseHelpQuery(query: string): string | null {
  const q = fold(query).trim();
  const m = HELP.exec(q);
  if (!m) return null;
  const rest = (m[1] ?? '').trim();
  // "?x" without a space is more likely a typo or a URL fragment than help.
  if (q.startsWith('?') && rest !== '' && !/^\?\s/.test(q)) return null;
  return rest;
}

/** Entries whose text or examples contain every word of the filter. */
export function filterHelp(entries: readonly HelpEntry[], filter: string): HelpEntry[] {
  const words = fold(filter).split(/\s+/).filter(Boolean);
  if (words.length === 0) return [...entries];
  return entries.filter((e) => {
    const haystack = fold([e.text, ...e.examples].join(' '));
    return words.every((w) => haystack.includes(w));
  });
}

/** The entries by group, in the order of HELP_GROUPS; empty groups are left out. */
export function groupHelp(entries: readonly HelpEntry[]): HelpSection[] {
  return HELP_GROUPS.map((group) => ({ group, entries: entries.filter((e) => e.group === group) })).filter(
    (s) => s.entries.length > 0,
  );
}

/**
 * Example queries for the rotating hint of the empty field: the first
 * example of each entry, taking turns between the groups so the hints do
 * not show five system questions in a row.
 */
export function hintExamples(entries: readonly HelpEntry[]): string[] {
  const queues = groupHelp(entries).map((s) => s.entries.flatMap((e) => e.examples.slice(0, 1)));
  const out: string[] = [];
  for (let i = 0; queues.some((q) => i < q.length); i++) {
    for (const q of queues) {
      const example = q[i];
      if (example) out.push(example);
    }
  }
  return out;
}
