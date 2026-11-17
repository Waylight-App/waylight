import { describe, expect, it } from 'vitest';
import { linkResult } from './identity';

describe('link result in the address (#527)', () => {
  it('reads the results the server sends and nothing else', () => {
    expect(linkResult(new URLSearchParams('settings=account&link=ok'))).toBe('ok');
    expect(linkResult(new URLSearchParams('link=taken'))).toBe('taken');
    expect(linkResult(new URLSearchParams('link=licence'))).toBe('licence');
    expect(linkResult(new URLSearchParams('link=<script>'))).toBeNull();
    expect(linkResult(new URLSearchParams(''))).toBeNull();
  });
});
