import { describe, expect, it } from 'vitest';
import raw from '../catalogue/feeds.json';
import {
  catalogueLanguages,
  type CatalogueEntry,
  sameFeedUrl,
  searchCatalogue,
  validEntry,
} from './catalogue';

const entries = raw as unknown[];

describe('feed catalogue (#289)', () => {
  it('has a valid data file without duplicate addresses', () => {
    expect(entries.length).toBeGreaterThan(20);
    for (const e of entries) expect(validEntry(e), JSON.stringify(e)).toBe(true);
    const urls = (entries as CatalogueEntry[]).map((e) => e.url.toLowerCase());
    expect(new Set(urls).size).toBe(urls.length);
  });

  it('searches by text and language, grouped by topic', () => {
    const list = entries as CatalogueEntry[];
    expect(catalogueLanguages(list)).toEqual(['de', 'en']);
    const de = searchCatalogue(list, '', 'de');
    expect(de.every((g) => g.entries.every((e) => e.language === 'de'))).toBe(true);
    expect(de[0]?.topic).toBe('news');
    const heise = searchCatalogue(list, 'HEISE', '');
    expect(heise).toEqual([{ topic: 'tech', entries: [expect.objectContaining({ title: 'heise online' })] }]);
    expect(searchCatalogue(list, 'no such feed', '')).toEqual([]);
  });

  it('rejects broken entries and compares addresses loosely', () => {
    expect(validEntry({ topic: 'news', language: 'de', title: 'x', url: 'http://plain.example/feed' })).toBe(
      false,
    );
    expect(validEntry({ topic: 'gossip', language: 'de', title: 'x', url: 'https://a.example' })).toBe(false);
    expect(validEntry({ topic: 'news', language: 'deu', title: 'x', url: 'https://a.example' })).toBe(false);
    expect(sameFeedUrl('https://Netzpolitik.org/feed/', 'http://netzpolitik.org/feed')).toBe(true);
    expect(sameFeedUrl('https://a.example/1', 'https://a.example/2')).toBe(false);
  });
});
