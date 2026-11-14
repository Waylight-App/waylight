import { describe, expect, it } from 'vitest';
import { age, parseConfig } from './mail';

describe('mail widget (#283)', () => {
  it('reads the account of a widget', () => {
    expect(parseConfig({})).toEqual({ account: '' });
    expect(parseConfig({ account: 'abc-1' })).toEqual({ account: 'abc-1' });
    expect(parseConfig({ account: '../x' })).toEqual({ account: '' });
  });
  it('describes the age of a message', () => {
    const now = new Date('2026-11-14T12:00:00Z');
    expect(age(new Date('2026-11-14T11:55:00Z'), now)).toBe('5 min');
    expect(age(new Date('2026-11-14T09:00:00Z'), now)).toBe('3 h');
    expect(age(new Date('2026-11-11T12:00:00Z'), now)).toBe('3 d');
    expect(age(null, now)).toBe('');
  });
});
