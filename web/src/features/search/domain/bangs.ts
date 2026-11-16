/**
 * Bang suggestions while typing "!" (#475). A curated list of the most used
 * DuckDuckGo bangs ships with the app, so nothing is asked from DuckDuckGo
 * while typing; the search engine resolves the bang as usual.
 */
export interface Bang {
  /** Without "!", lower case. */
  bang: string;
  name: string;
  /** From Settings → Search rather than the built-in list. */
  own?: boolean;
}

/** Engines that understand DuckDuckGo-style bangs. */
export const BANG_ENGINES = new Set(['duckduckgo', 'brave']);

export const BANGS: Bang[] = [
  { bang: 'g', name: 'Google' },
  { bang: 'gi', name: 'Google Images' },
  { bang: 'gm', name: 'Google Maps' },
  { bang: 'gh', name: 'GitHub' },
  { bang: 'gist', name: 'GitHub Gist' },
  { bang: 'w', name: 'Wikipedia' },
  { bang: 'wde', name: 'Wikipedia (Deutsch)' },
  { bang: 'yt', name: 'YouTube' },
  { bang: 'a', name: 'Amazon' },
  { bang: 'ade', name: 'Amazon.de' },
  { bang: 'ebay', name: 'eBay' },
  { bang: 'r', name: 'Reddit' },
  { bang: 'so', name: 'Stack Overflow' },
  { bang: 'mdn', name: 'MDN Web Docs' },
  { bang: 'npm', name: 'npm' },
  { bang: 'pypi', name: 'PyPI' },
  { bang: 'godoc', name: 'Go packages' },
  { bang: 'dh', name: 'Docker Hub' },
  { bang: 'aw', name: 'ArchWiki' },
  { bang: 'imdb', name: 'IMDb' },
  { bang: 'osm', name: 'OpenStreetMap' },
  { bang: 'maps', name: 'Maps' },
  { bang: 'img', name: 'Images' },
  { bang: 'news', name: 'News' },
  { bang: 'dict', name: 'Dictionary' },
  { bang: 'leo', name: 'LEO (Englisch–Deutsch)' },
  { bang: 'dcc', name: 'dict.cc' },
  { bang: 'deepl', name: 'DeepL' },
  { bang: 'gt', name: 'Google Translate' },
  { bang: 'wa', name: 'Wolfram Alpha' },
  { bang: 'hn', name: 'Hacker News' },
  { bang: 'x', name: 'X' },
  { bang: 'mastodon', name: 'Mastodon' },
  { bang: 'spotify', name: 'Spotify' },
  { bang: 'idealo', name: 'idealo' },
  { bang: 'geizhals', name: 'Geizhals' },
  { bang: 'db', name: 'Deutsche Bahn' },
];

/** Most suggestions shown at once. */
export const MAX_BANGS = 8;

/**
 * The bang being typed: "!g" or "!gh" as the only word, without the
 * search terms that follow. Null when the text is not a lone bang.
 */
export function typedBang(query: string): string | null {
  const m = /^!([\p{L}\p{N}_-]*)$/u.exec(query.trimStart());
  return m ? (m[1] ?? '').toLowerCase() : null;
}

/**
 * Bangs that start with what is typed: own shortcuts first (they always
 * work), then the built-in list when the engine supports bangs. An exact
 * match comes first in each part.
 */
export function suggestBangs(query: string, own: readonly Bang[], engineSupportsBangs: boolean): Bang[] {
  const typed = typedBang(query);
  if (typed === null) return [];
  const pick = (list: readonly Bang[]) =>
    list
      .filter((b) => b.bang.startsWith(typed))
      .sort((a, b) => Number(b.bang === typed) - Number(a.bang === typed) || a.bang.length - b.bang.length);
  const mine = pick(own.map((b) => ({ ...b, own: true })));
  const taken = new Set(mine.map((b) => b.bang));
  const builtIn = engineSupportsBangs ? pick(BANGS).filter((b) => !taken.has(b.bang)) : [];
  return [...mine, ...builtIn].slice(0, MAX_BANGS);
}
