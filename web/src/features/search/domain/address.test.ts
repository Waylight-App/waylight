import { describe, expect, it } from 'vitest';
import { detectAddress, displayAddress } from './address';
import { BANGS, suggestBangs, typedBang } from './bangs';

describe('address detection (#475)', () => {
  it.each([
    ['github.com', 'https://github.com/'],
    ['GitHub.com/itadori26', 'https://github.com/itadori26'],
    ['docs.example.org/a?b=1', 'https://docs.example.org/a?b=1'],
    ['192.168.178.1', 'http://192.168.178.1/'],
    ['10.0.0.5:8080/admin', 'http://10.0.0.5:8080/admin'],
    ['nas.local:5001', 'http://nas.local:5001/'],
    ['pi.lan', 'http://pi.lan/'],
    ['router.home.arpa', 'http://router.home.arpa/'],
    ['fritz.box', 'http://fritz.box/'],
    ['my.fritz.box', 'http://my.fritz.box/'],
    ['localhost:3000', 'http://localhost:3000/'],
    ['http://router', 'http://router/'],
    ['https://example.org/x', 'https://example.org/x'],
  ])('opens %s', (input, url) => {
    expect(detectAddress(input)).toBe(url);
  });

  it.each([
    '5.5',
    '3.14',
    'v1.2.3',
    '1.2.3',
    'node.js tutorial',
    'node.js',
    'readme.md',
    'github',
    'github com',
    '999.1.1.1',
    'example.com is down',
    'mailto:me@example.org',
    'a..com',
    '',
  ])('searches for %s', (input) => {
    expect(detectAddress(input)).toBeNull();
  });

  it('shows the address without scheme', () => {
    expect(displayAddress('https://github.com/')).toBe('github.com');
    expect(displayAddress('http://nas.local:5001/')).toBe('nas.local:5001');
  });
});

describe('bang suggestions (#475)', () => {
  it('lists bangs for "!" and filters by the letters typed', () => {
    expect(suggestBangs('!', [], true)).toHaveLength(8);
    const g = suggestBangs('!g', [], true).map((b) => b.bang);
    expect(g[0]).toBe('g');
    expect(g).toContain('gh');
    expect(g).toContain('gi');
    expect(g.every((b) => b.startsWith('g'))).toBe(true);
  });

  it('puts own shortcuts first and marks them', () => {
    const own = [{ bang: 'gh', name: 'My GitHub' }];
    const list = suggestBangs('!gh', own, true);
    expect(list[0]).toEqual({ bang: 'gh', name: 'My GitHub', own: true });
    expect(list.filter((b) => b.bang === 'gh')).toHaveLength(1);
  });

  it('offers only own shortcuts when the engine has no bangs', () => {
    expect(suggestBangs('!g', [{ bang: 'gl', name: 'GitLab' }], false)).toEqual([
      { bang: 'gl', name: 'GitLab', own: true },
    ]);
  });

  it('stops suggesting once the search terms start', () => {
    expect(typedBang('!gh')).toBe('gh');
    expect(typedBang('!gh trisse')).toBeNull();
    expect(typedBang('hello')).toBeNull();
    expect(suggestBangs('!w Berlin', [], true)).toEqual([]);
  });

  it('has no duplicate bangs in the built-in list', () => {
    expect(new Set(BANGS.map((b) => b.bang)).size).toBe(BANGS.length);
  });
});
