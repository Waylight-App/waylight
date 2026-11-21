import type { MobileSize } from './mobile';

/**
 * Widget sizes on the web grid (#595). They are the phone's sizes (#585),
 * so a widget means the same thing on both: small and strip show the
 * compact view, medium and large the full view. A widget without a size is
 * a freely sized tile as before and always shows its full view.
 */
export type WidgetSize = MobileSize;

/** The sizes in the order the size menu offers them, with their cells on the 12-column grid. */
export const WIDGET_SIZES: readonly { size: WidgetSize; w: number; h: number }[] = [
  { size: 'small', w: 3, h: 2 },
  { size: 'medium', w: 4, h: 4 },
  { size: 'large', w: 6, h: 6 },
  { size: 'strip', w: 6, h: 1 },
];

/** Smallest content box, in pixels, in which a full view still makes sense. */
export const FULL_VIEW_MIN = { width: 200, height: 120 } as const;

/** Parses a size from the API; anything unknown is a freely sized tile. */
export function parseSize(value: string | undefined): WidgetSize | undefined {
  return WIDGET_SIZES.find((s) => s.size === value)?.size;
}

/** Grid cells of a size. */
export function sizeRect(size: WidgetSize): { w: number; h: number } {
  const s = WIDGET_SIZES.find((x) => x.size === size) ?? WIDGET_SIZES[0];
  return { w: s?.w ?? 3, h: s?.h ?? 2 };
}

/** Whether a size shows the compact view. */
export function isCompactSize(size: WidgetSize | undefined): boolean {
  return size === 'small' || size === 'strip';
}

/** The size whose cells are closest to w × h, so dragging the edge of a sized tile snaps. */
export function snapSize(w: number, h: number): WidgetSize {
  let best = WIDGET_SIZES[0];
  let bestDistance = Infinity;
  for (const s of WIDGET_SIZES) {
    const d = Math.abs(s.w - w) + Math.abs(s.h - h);
    if (d < bestDistance) [best, bestDistance] = [s, d];
  }
  return best?.size ?? 'small';
}

/**
 * The order keyboard resizing walks, from one line to the biggest tile.
 * Strip and small cover the same area, so the order is spelt out.
 */
const BY_AREA: readonly WidgetSize[] = ['strip', 'small', 'medium', 'large'];

/** The next larger (grow) or smaller size, or null at either end. */
export function stepSize(size: WidgetSize, grow: boolean): WidgetSize | null {
  const next = BY_AREA[BY_AREA.indexOf(size) + (grow ? 1 : -1)];
  return next ?? null;
}

/** How a widget is shown on the dashboard. */
export type WidgetView = 'full' | 'compact';

/**
 * Decides between the compact and the full view. A widget without a
 * compact view always shows its full view. Small and strip tiles are
 * compact; medium and large tiles fall back to compact when the window
 * makes them too small for their full view (cramped). Freely sized tiles
 * keep their full view, so existing dashboards look as before.
 */
export function widgetView(o: {
  size: WidgetSize | undefined;
  hasCompact: boolean;
  cramped: boolean;
}): WidgetView {
  if (!o.hasCompact || o.size === undefined) return 'full';
  if (isCompactSize(o.size) || o.cramped) return 'compact';
  return 'full';
}

/** Whether a content box is too small for a full view. */
export function isCramped(width: number, height: number): boolean {
  return width > 0 && height > 0 && (width < FULL_VIEW_MIN.width || height < FULL_VIEW_MIN.height);
}
