import { describe, expect, it } from 'vitest';
import { appCallback, parseAppHandover } from './appHandover';

const challenge = 'E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM';

describe('parseAppHandover', () => {
  it('accepts an S256 challenge and a state', () => {
    expect(parseAppHandover(`?challenge=${challenge}&state=abc-1`)).toEqual({ challenge, state: 'abc-1' });
  });

  it('rejects missing or malformed values', () => {
    expect(parseAppHandover('?state=abc')).toBeNull();
    expect(parseAppHandover(`?challenge=short&state=abc`)).toBeNull();
    expect(parseAppHandover(`?challenge=${challenge}&state=a%20b`)).toBeNull();
  });
});

describe('appCallback', () => {
  it('only ever links to the app', () => {
    const url = new URL(appCallback({ challenge, state: 's1' }, { code: 'c&d' }));
    expect(url.protocol).toBe('waylight:');
    expect(url.searchParams.get('code')).toBe('c&d');
    expect(url.searchParams.get('state')).toBe('s1');
  });

  it('reports a cancelled sign-in', () => {
    const url = new URL(appCallback({ challenge, state: 's1' }, { error: 'access_denied' }));
    expect(url.searchParams.get('error')).toBe('access_denied');
  });
});
