import type { WidgetCategory } from './catalog';
import type { Widget } from './dashboard';

/**
 * The zones of a zoned tab (#619) in display order: what matters today,
 * the viewer's own things, and the state of the server.
 */
export const ZONES = ['today', 'mine', 'system'] as const;
export type Zone = (typeof ZONES)[number];

/** Parses a zone from the API; unknown or empty values mean "by kind". */
export function parseZone(value: string | undefined): Zone | undefined {
  return ZONES.find((z) => z === value);
}

/** Kinds that answer "what is today like": time, weather, next event, waste, to-dos. */
const TODAY_KINDS: ReadonlySet<string> = new Set(['clock', 'weather', 'calendar', 'waste', 'todos']);

/** Picker categories whose widgets watch the server or the network. */
const SYSTEM_CATEGORIES: ReadonlySet<WidgetCategory> = new Set(['server', 'network']);

/** The zone a widget kind falls into until someone moves it. */
export function defaultZone(kind: string, category: WidgetCategory | undefined): Zone {
  if (TODAY_KINDS.has(kind)) return 'today';
  if (category && SYSTEM_CATEGORIES.has(category)) return 'system';
  return 'mine';
}

/** The zone a widget is shown in: where it was moved to, else its default. */
export function zoneOf(widget: Widget, category: WidgetCategory | undefined): Zone {
  return widget.zone ?? defaultZone(widget.kind, category);
}

/** One zone with its widgets in reading order. */
export interface ZoneGroup {
  zone: Zone;
  widgets: Widget[];
}

/**
 * Splits widgets into zones. Within a zone the free grid's reading order
 * (top to bottom, left to right) is kept, so a tab looks familiar when
 * zones are switched on. The System zone is left out for viewers who are
 * not admins, and empty zones are dropped unless `keepEmpty` is set (edit
 * mode, where an empty zone is still a place to move widgets to).
 */
export function groupByZone(
  widgets: readonly Widget[],
  categoryOf: (kind: string) => WidgetCategory | undefined,
  opts: { system: boolean; keepEmpty?: boolean },
): ZoneGroup[] {
  const sorted = [...widgets].sort((a, b) => a.y - b.y || a.x - b.x);
  return ZONES.filter((zone) => zone !== 'system' || opts.system)
    .map((zone) => ({ zone, widgets: sorted.filter((w) => zoneOf(w, categoryOf(w.kind)) === zone) }))
    .filter((g) => opts.keepEmpty || g.widgets.length > 0);
}
