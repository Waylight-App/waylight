import { MAX_ROWS, type Rect, sortByPosition } from './layout';

/**
 * Phone layout of a tab (#92): one column, widgets in an order of their
 * own, each with a minimum height and an optional "hidden on phones" flag.
 * The backend stores and resolves it; these functions mirror the rules so
 * local edits (added or deleted widgets) show up without a reload.
 */
export interface MobileItem {
  id: string;
  hidden: boolean;
  /** Minimum height in grid rows, used while no size is set. */
  rows: number;
  /** Size on the phone grid of four columns (#585); unset means full width. */
  size?: MobileSize;
}

/** Sizes of the phone grid, as on phone home screens (#585). */
export type MobileSize = 'small' | 'medium' | 'large' | 'strip';

/** The sizes in the order the layout editor offers them, with their cells. */
export const MOBILE_SIZES: readonly { size: MobileSize; cols: number; rows: number }[] = [
  { size: 'small', cols: 2, rows: 2 },
  { size: 'medium', cols: 4, rows: 2 },
  { size: 'large', cols: 4, rows: 4 },
  { size: 'strip', cols: 4, rows: 1 },
];

/** Columns and rows a size spans; undefined for a full-width widget. */
export function sizeCells(size: MobileSize | undefined): { cols: number; rows: number } | undefined {
  return MOBILE_SIZES.find((s) => s.size === size);
}

/** Whether any widget of the layout has a size, so the grid applies. */
export function usesGrid(items: readonly MobileItem[]): boolean {
  return items.some((i) => sizeCells(i.size) !== undefined);
}

/** Width below which the tabs become a bottom navigation (#771); matches the CSS breakpoint. */
export const BOTTOM_NAV_QUERY = '(max-width: 639px)';

/** Width below which the phone layout is used; matches the CSS breakpoint. */
export const PHONE_QUERY = '(max-width: 720px)';

const clampRows = (rows: number) => Math.min(Math.max(Math.round(rows), 1), MAX_ROWS);

/** Builds the phone layout from the desktop grid: reading order, desktop height. */
export function deriveMobile(widgets: readonly (Rect & { id: string })[]): MobileItem[] {
  return sortByPosition(widgets).map((w) => ({ id: w.id, hidden: false, rows: w.h }));
}

/**
 * Applies a stored layout to the current widgets: entries of deleted
 * widgets disappear, new widgets are appended in reading order.
 */
export function resolveMobile(
  stored: readonly MobileItem[],
  widgets: readonly (Rect & { id: string })[],
): MobileItem[] {
  const present = new Set(widgets.map((w) => w.id));
  const seen = new Set<string>();
  const out: MobileItem[] = [];
  for (const e of stored) {
    if (!present.has(e.id) || seen.has(e.id)) continue;
    seen.add(e.id);
    // A size from a newer server falls back to full width.
    const size = sizeCells(e.size) ? e.size : undefined;
    out.push({ id: e.id, hidden: e.hidden, rows: clampRows(e.rows), ...(size ? { size } : {}) });
  }
  for (const item of deriveMobile(widgets)) if (!seen.has(item.id)) out.push(item);
  return out;
}

/** Moves an item by `delta` places; returns null at the edges. */
export function moveMobile(items: readonly MobileItem[], id: string, delta: number): MobileItem[] | null {
  const from = items.findIndex((i) => i.id === id);
  const to = from + delta;
  if (from < 0 || to < 0 || to >= items.length) return null;
  const out = [...items];
  const [moved] = out.splice(from, 1);
  if (moved) out.splice(to, 0, moved);
  return out;
}

/** Changes one item (height clamped to the grid limits). */
export function updateMobile(
  items: readonly MobileItem[],
  id: string,
  change: Partial<Pick<MobileItem, 'hidden' | 'rows' | 'size'>>,
): MobileItem[] {
  return items.map((i) => {
    if (i.id !== id) return i;
    const next: MobileItem = { ...i, ...change, rows: clampRows(change.rows ?? i.rows) };
    if ('size' in change && !change.size) delete next.size;
    return next;
  });
}
