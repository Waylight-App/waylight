import { describe, expect, it } from 'vitest';
import { serverMessages } from './generated/server-messages';
import { serverDe } from './serverDe';
import { plugin } from './plugin';

const placeholders = (s: string) => [...s.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort();

describe('ee/ backend messages (#632)', () => {
  it('translates every message with the same placeholders', () => {
    for (const [key, english] of Object.entries(serverMessages)) {
      const german = serverDe[key as keyof typeof serverDe];
      expect(german, key).toBeTruthy();
      expect(placeholders(german), key).toEqual(placeholders(english));
    }
  });

  it('hands the translations to the core', () => {
    expect(plugin.serverMessages?.de).toBe(serverDe);
  });

  it('gives the presence widget a compact view for its group', () => {
    const presence = plugin.widgets.find((w) => w.kind === 'presence');
    expect(presence?.compact?.({ group: 'g1' })).toEqual({ module: 'presence', params: { group: 'g1' } });
    expect(presence?.compact?.({})).toEqual({ module: 'presence', params: {} });
    expect(presence?.compactExample?.de).toBeTruthy();
  });
});
