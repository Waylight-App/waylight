/** One source of the built-in feed catalogue (#289, catalogue/feeds.json). */
export interface CatalogueEntry {
  topic: CatalogueTopic;
  language: string;
  title: string;
  url: string;
}

export const CATALOGUE_TOPICS = ['news', 'tech', 'selfhosting', 'science', 'sports', 'fun'] as const;
export type CatalogueTopic = (typeof CATALOGUE_TOPICS)[number];

/** Checks one entry of the data file (format documented in its README). */
export function validEntry(e: unknown): e is CatalogueEntry {
  if (typeof e !== 'object' || e === null) return false;
  const r = e as Record<string, unknown>;
  return (
    CATALOGUE_TOPICS.includes(r['topic'] as CatalogueTopic) &&
    typeof r['language'] === 'string' &&
    /^[a-z]{2}$/.test(r['language']) &&
    typeof r['title'] === 'string' &&
    r['title'].trim() !== '' &&
    typeof r['url'] === 'string' &&
    /^https:\/\/[^\s]+$/.test(r['url'])
  );
}

/** Languages offered by the catalogue, in a stable order. */
export function catalogueLanguages(entries: readonly CatalogueEntry[]): string[] {
  return [...new Set(entries.map((e) => e.language))].sort();
}

/**
 * Entries matching a search text (title or address) and a language ('' =
 * all), grouped by topic in topic order; empty topics are left out.
 */
export function searchCatalogue(
  entries: readonly CatalogueEntry[],
  query: string,
  language: string,
): { topic: CatalogueTopic; entries: CatalogueEntry[] }[] {
  const q = query.trim().toLowerCase();
  const hits = entries.filter(
    (e) =>
      (!language || e.language === language) &&
      (!q || e.title.toLowerCase().includes(q) || e.url.toLowerCase().includes(q)),
  );
  return CATALOGUE_TOPICS.map((topic) => ({ topic, entries: hits.filter((e) => e.topic === topic) })).filter(
    (g) => g.entries.length > 0,
  );
}

/** Normalised address for "already subscribed" (scheme, host case, trailing slash). */
export function sameFeedUrl(a: string, b: string): boolean {
  const norm = (u: string) =>
    u
      .trim()
      .toLowerCase()
      .replace(/^http:/, 'https:')
      .replace(/\/+$/, '');
  return norm(a) === norm(b);
}
