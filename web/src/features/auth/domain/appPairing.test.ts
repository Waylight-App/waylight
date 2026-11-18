import { describe, expect, it } from 'vitest';
import { pairingLink } from './appPairing';
import { describeSession } from './session';

describe('pairingLink', () => {
  it('carries the server address and the code for the app', () => {
    const link = new URL(pairingLink('https://home.example.org', 'a+b/c'));
    expect(link.protocol).toBe('waylight:');
    expect(link.searchParams.get('server')).toBe('https://home.example.org');
    expect(link.searchParams.get('code')).toBe('a+b/c');
  });
});

describe('describeSession', () => {
  it('names app sessions after their device', () => {
    expect(describeSession({ device: 'Pixel 9', userAgent: 'Dart/3.9' })).toContain('Pixel 9');
  });

  it('names browser sessions after the browser', () => {
    expect(
      describeSession({ device: '', userAgent: 'Mozilla/5.0 (X11; Linux x86_64) Firefox/140.0' }),
    ).toContain('Firefox');
  });
});
