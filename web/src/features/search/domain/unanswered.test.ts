import { describe, expect, it } from 'vitest';
import { addUnanswered, groupUnanswered, isRecordable, MAX_UNANSWERED, parseUnanswered } from './unanswered';

const T0 = Date.UTC(2026, 10, 16);

describe('which queries may be remembered (#481)', () => {
  it('keeps ordinary questions', () => {
    for (const q of ['zug nach köln', 'Wann ist Müllabfuhr', 'printer status', 'kfz steuer'])
      expect(isRecordable(q)).toBe(true);
  });

  it('leaves out bangs', () => {
    for (const q of ['!g cats', 'cats !w', '!yt']) expect(isRecordable(q)).toBe(false);
    expect(isRecordable('wow! great')).toBe(true);
  });

  it('leaves out what looks like a secret or personal data', () => {
    for (const q of [
      'hunter2Password!xyz',
      'ghp_a1B2c3D4e5F6g7H8i9J0k1L2m3N4',
      'token 3f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c',
      'aB3dE5gH7jK9mN',
      'max@example.com',
      'https://intranet/secret?id=1',
      'www.example.com',
      'DE89 3704 0044 0532 0130 00',
      '0151 2345678',
    ])
      expect(isRecordable(q), q).toBe(false);
  });

  it('leaves out empty and very long input', () => {
    expect(isRecordable(' ')).toBe(false);
    expect(isRecordable('a'.repeat(5) + ' word'.repeat(20))).toBe(false);
  });
});

describe('the list of unanswered queries (#481)', () => {
  it('counts repeated queries case- and accent-insensitively', () => {
    let e = addUnanswered([], 'Zug nach Köln', T0);
    e = addUnanswered(e, 'zug  nach koln', T0 + 1);
    expect(e).toEqual([{ query: 'zug nach koln', count: 2, last: T0 + 1 }]);
  });

  it('keeps the newest entries up to the limit and ignores filtered ones', () => {
    let e: ReturnType<typeof addUnanswered> = [];
    for (let i = 0; i < MAX_UNANSWERED + 5; i++) e = addUnanswered(e, `frage ${i}`, T0 + i);
    expect(e).toHaveLength(MAX_UNANSWERED);
    expect(e[0]?.query).toBe(`frage ${MAX_UNANSWERED + 4}`);
    expect(addUnanswered(e, '!g x', T0)).toEqual(e);
  });

  it('groups by the first meaningful word, most asked first', () => {
    let e = addUnanswered([], 'zug nach köln', T0);
    e = addUnanswered(e, 'zug morgen', T0);
    e = addUnanswered(e, 'zug morgen', T0);
    e = addUnanswered(e, 'wie ist der pegel', T0);
    expect(groupUnanswered(e).map((g) => [g.word, g.count, g.queries.length])).toEqual([
      ['zug', 3, 2],
      ['pegel', 1, 1],
    ]);
  });

  it('validates stored data', () => {
    expect(parseUnanswered('x')).toEqual([]);
    expect(
      parseUnanswered([{ query: 'a', count: 1, last: T0 }, { query: 'b', count: 0, last: T0 }, null]),
    ).toEqual([{ query: 'a', count: 1, last: T0 }]);
  });
});
