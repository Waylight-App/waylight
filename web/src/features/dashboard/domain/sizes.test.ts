import { describe, expect, it } from 'vitest';
import { isCompactSize, isCramped, parseSize, sizeRect, snapSize, stepSize, widgetView } from './sizes';

describe('widget sizes (#595)', () => {
  it('parses known sizes and treats anything else as a free tile', () => {
    expect(parseSize('strip')).toBe('strip');
    expect(parseSize('')).toBeUndefined();
    expect(parseSize('huge')).toBeUndefined();
    expect(parseSize(undefined)).toBeUndefined();
  });

  it('gives every size its cells on the 12-column grid', () => {
    expect(sizeRect('small')).toEqual({ w: 3, h: 2 });
    expect(sizeRect('strip')).toEqual({ w: 6, h: 1 });
    expect(isCompactSize('small')).toBe(true);
    expect(isCompactSize('large')).toBe(false);
    expect(isCompactSize(undefined)).toBe(false);
  });

  it('snaps a dragged edge to the nearest size', () => {
    expect(snapSize(3, 2)).toBe('small');
    expect(snapSize(5, 4)).toBe('medium');
    expect(snapSize(7, 7)).toBe('large');
    expect(snapSize(8, 1)).toBe('strip');
  });

  it('steps through the sizes by area for keyboard resizing', () => {
    expect(stepSize('strip', true)).toBe('small');
    expect(stepSize('small', true)).toBe('medium');
    expect(stepSize('large', true)).toBeNull();
    expect(stepSize('strip', false)).toBeNull();
  });

  it('decides the view from size, compact support and room', () => {
    const view = (size: Parameters<typeof widgetView>[0]['size'], hasCompact = true, cramped = false) =>
      widgetView({ size, hasCompact, cramped });
    expect(view('small')).toBe('compact');
    expect(view('strip')).toBe('compact');
    expect(view('medium')).toBe('full');
    expect(view('large', true, true)).toBe('compact');
    // Widgets without a compact view and free tiles keep their full view.
    expect(view('small', false)).toBe('full');
    expect(view(undefined, true, true)).toBe('full');
  });

  it('calls a box cramped below the minimum of a full view', () => {
    expect(isCramped(150, 300)).toBe(true);
    expect(isCramped(300, 80)).toBe(true);
    expect(isCramped(300, 200)).toBe(false);
    // Not measured yet (hidden tab): no fallback.
    expect(isCramped(0, 0)).toBe(false);
  });
});
