import { describe, expect, it } from 'vitest';
import { newId } from './uuid';

describe('newId (#310)', () => {
  it('works without crypto.randomUUID (plain HTTP on a LAN address)', () => {
    const insecure = {
      getRandomValues: <T extends ArrayBufferView | null>(a: T) => crypto.getRandomValues(a),
    };
    const ids = new Set(Array.from({ length: 50 }, () => newId(insecure)));
    expect(ids.size).toBe(50);
    for (const id of ids)
      expect(id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  });
  it('uses crypto.randomUUID where it exists', () => {
    expect(newId({ getRandomValues: (a) => a, randomUUID: () => 'x-y' })).toBe('x-y');
  });
});
