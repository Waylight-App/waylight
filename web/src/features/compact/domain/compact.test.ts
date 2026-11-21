import { describe, expect, it } from 'vitest';
import { compactText, isAlarming, parseLevel } from './compact';

describe('compact state (#595)', () => {
  it('parses levels and treats unknown ones as none', () => {
    expect(parseLevel('problem')).toBe('problem');
    expect(parseLevel('critical')).toBe('none');
  });

  it('colours only attention and problem', () => {
    expect(isAlarming('ok')).toBe(false);
    expect(isAlarming('none')).toBe(false);
    expect(isAlarming('attention')).toBe(true);
    expect(isAlarming('problem')).toBe(true);
  });

  it('falls back to the English text the server filled in', () => {
    expect(compactText({ key: 'srv_unknown', text: '3 open', params: { count: '3' } })).toBe('3 open');
  });
});
