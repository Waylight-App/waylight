import { describe, expect, it } from 'vitest';
import { simpleLayout } from './simple';

describe('simpleLayout (#681)', () => {
  const w = (id: string, x: number, y: number, h = 3) => ({ id, x, y, w: 3, h });

  it('lists the widgets in reading order, compact ones as small tiles', () => {
    const items = simpleLayout([w('b', 6, 0), w('c', 0, 4, 5), w('a', 0, 0)], (x) => x.id !== 'c');
    expect(items.map((i) => [i.widget.id, i.size, i.rows])).toEqual([
      ['a', 'small', 2],
      ['b', 'small', 2],
      ['c', undefined, 5],
    ]);
  });
});
