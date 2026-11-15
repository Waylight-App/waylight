/** Wiki software that can be searched (#321, #322, #323). */
export type WikiKind = 'bookstack' | 'outline' | 'wikijs';

export const WIKI_KINDS: readonly WikiKind[] = ['bookstack', 'outline', 'wikijs'];

/** Display names of the wiki software. */
export const WIKI_KIND_NAMES: Record<WikiKind, string> = {
  bookstack: 'BookStack',
  outline: 'Outline',
  wikijs: 'Wiki.js',
};

/** A connected wiki; the token never leaves the server. */
export interface Wiki {
  id: string;
  name: string;
  kind: WikiKind;
  url: string;
  hasToken: boolean;
}

/** Input of a wiki; an empty token keeps the stored one. */
export interface WikiInput {
  name: string;
  kind: WikiKind;
  url: string;
  token: string;
}

/** A page a wiki found. */
export interface WikiPage {
  title: string;
  excerpt: string;
  url: string;
  space: string;
  wiki: string;
  updated: Date | null;
}

/** Merged answer of all wikis; failed names the wikis that did not answer. */
export interface WikiSearch {
  pages: WikiPage[];
  failed: string[];
}

/** Shortest query sent to the wikis; the server enforces the same. */
export const MIN_QUERY_LENGTH = 2;

/**
 * "w handbook" or "wiki handbook" searches the wikis only (#321). Returns
 * the query without the prefix, or null when the query has none.
 */
export function wikiOnlyQuery(query: string): string | null {
  const m = /^(?:w|wiki)\s+(.+)$/i.exec(query.trim());
  return m ? (m[1] ?? '').trim() : null;
}

/** "Space · Wiki" or whichever of the two is known, for the result line. */
export function pageDetail(p: Pick<WikiPage, 'space' | 'wiki' | 'excerpt'>): string {
  return [[p.space, p.wiki].filter(Boolean).join(' · '), p.excerpt].filter(Boolean).join(' — ');
}

/** The token field explains what each wiki expects. */
export const TOKEN_HINT: Record<
  WikiKind,
  'wiki.token_hint_bookstack' | 'wiki.token_hint_outline' | 'wiki.token_hint_wikijs'
> = {
  bookstack: 'wiki.token_hint_bookstack',
  outline: 'wiki.token_hint_outline',
  wikijs: 'wiki.token_hint_wikijs',
};
