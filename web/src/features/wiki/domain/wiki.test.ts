import { describe, expect, it } from 'vitest';
import { pageDetail, wikiOnlyQuery } from './wiki';

describe('wiki domain', () => {
  it('recognises the wiki-only prefix', () => {
    expect(wikiOnlyQuery('w vpn setup')).toBe('vpn setup');
    expect(wikiOnlyQuery('Wiki  printer')).toBe('printer');
    expect(wikiOnlyQuery('wetter')).toBeNull();
    expect(wikiOnlyQuery('w')).toBeNull();
  });

  it('builds the result line', () => {
    expect(pageDetail({ space: 'IT', wiki: 'Handbook', excerpt: 'How to' })).toBe('IT · Handbook — How to');
    expect(pageDetail({ space: '', wiki: 'Handbook', excerpt: '' })).toBe('Handbook');
  });
});
