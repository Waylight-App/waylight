import { type Rect, sortByPosition } from './layout';
import type { WidgetSize } from './sizes';

/** One widget of the simple view (#681). */
export interface SimpleItem<T> {
  widget: T;
  /** Small for a compact view; undefined keeps the full view. */
  size: WidgetSize | undefined;
  /** Height of a full view, in grid rows. */
  rows: number;
}

/**
 * The simple view (#681): one column in reading order of the grid. A widget
 * with a compact view shows only that; one without keeps its full view at
 * its own height, because it has nothing shorter to show.
 */
export function simpleLayout<T extends Rect>(
  widgets: readonly T[],
  hasCompact: (widget: T) => boolean,
): SimpleItem<T>[] {
  return sortByPosition(widgets).map((widget) =>
    hasCompact(widget) ? { widget, size: 'small', rows: 2 } : { widget, size: undefined, rows: widget.h },
  );
}
