import { describe, expect, it } from 'vitest';
import { validateNote } from './board';

describe('family board domain', () => {
  it('validates text and expiry', () => {
    const ok = { text: 'Dentist on Friday', pinned: false, expiresOn: '' };
    expect(validateNote(ok, '2026-11-23')).toBeNull();
    expect(validateNote({ ...ok, expiresOn: '2026-11-23' }, '2026-11-23')).toBeNull();
    expect(validateNote({ ...ok, expiresOn: '2026-11-22' }, '2026-11-23')).toBe('board.error.expiry');
    expect(validateNote({ ...ok, text: ' ' }, '2026-11-23')).toBe('board.error.text');
    expect(validateNote({ ...ok, text: 'x'.repeat(501) }, '2026-11-23')).toBe('board.error.text');
  });
});
