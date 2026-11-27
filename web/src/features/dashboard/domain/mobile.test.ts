import { describe, expect, it } from 'vitest';
import { deriveMobile, moveMobile, resolveMobile, sizeCells, updateMobile, usesGrid } from './mobile';

const w = (id: string, x: number, y: number, h = 2) => ({ id, x, y, w: 4, h });

describe('phone layout (#92)', () => {
  it('derives reading order and desktop heights', () => {
    expect(deriveMobile([w('c', 0, 4), w('b', 8, 0, 3), w('a', 0, 0, 5)])).toEqual([
      { id: 'a', hidden: false, rows: 5 },
      { id: 'b', hidden: false, rows: 3 },
      { id: 'c', hidden: false, rows: 2 },
    ]);
  });

  it('keeps the stored order, drops deleted and appends new widgets', () => {
    const stored = [
      { id: 'b', hidden: true, rows: 99 },
      { id: 'gone', hidden: false, rows: 2 },
      { id: 'a', hidden: false, rows: 1 },
      { id: 'a', hidden: false, rows: 7 },
    ];
    expect(resolveMobile(stored, [w('a', 0, 0), w('b', 4, 0), w('new', 0, 9, 3)])).toEqual([
      { id: 'b', hidden: true, rows: 24 },
      { id: 'a', hidden: false, rows: 1 },
      { id: 'new', hidden: false, rows: 3 },
    ]);
  });

  it('moves, hides and resizes items', () => {
    const items = deriveMobile([w('a', 0, 0), w('b', 0, 2), w('c', 0, 4)]);
    expect(moveMobile(items, 'c', -2)?.map((i) => i.id)).toEqual(['c', 'a', 'b']);
    expect(moveMobile(items, 'a', -1)).toBeNull();
    expect(moveMobile(items, 'c', 1)).toBeNull();
    expect(updateMobile(items, 'b', { hidden: true })[1]).toEqual({ id: 'b', hidden: true, rows: 2 });
    expect(updateMobile(items, 'b', { rows: 0 })[1]?.rows).toBe(1);
    expect(updateMobile(items, 'b', { rows: 30 })[1]?.rows).toBe(24);
  });
});

describe('sizes on the phone grid (#585)', () => {
  const widgets = [
    { id: 'a', x: 0, y: 0, w: 4, h: 2 },
    { id: 'b', x: 4, y: 0, w: 4, h: 3 },
  ];

  it('knows the cells of each size', () => {
    expect(sizeCells('small')).toEqual({ size: 'small', cols: 2, rows: 2 });
    expect(sizeCells('strip')?.rows).toBe(1);
    expect(sizeCells(undefined)).toBeUndefined();
  });

  it('keeps known sizes and drops unknown ones', () => {
    const stored = [
      { id: 'a', hidden: false, rows: 2, size: 'large' as const },
      { id: 'b', hidden: false, rows: 3, size: 'giant' as never },
    ];
    const out = resolveMobile(stored, widgets);
    expect(out[0]?.size).toBe('large');
    expect(out[1]).not.toHaveProperty('size');
    expect(usesGrid(out)).toBe(true);
    expect(usesGrid(deriveMobile(widgets))).toBe(false);
  });

  it('sets and clears a size', () => {
    const sized = updateMobile(deriveMobile(widgets), 'a', { size: 'medium' });
    expect(sized[0]?.size).toBe('medium');
    expect(updateMobile(sized, 'a', { size: undefined })[0]).not.toHaveProperty('size');
  });

  it('keeps favourites set in the app (#844)', () => {
    const stored = [
      { id: 'a', hidden: false, rows: 2, favourite: true },
      { id: 'b', hidden: false, rows: 3 },
    ];
    const out = resolveMobile(stored, widgets);
    expect(out[0]?.favourite).toBe(true);
    expect(out[1]).not.toHaveProperty('favourite');
    expect(updateMobile(out, 'a', { rows: 4 })[0]?.favourite).toBe(true);
  });
});
