/** Part of the day that decides the greeting (#167). */
export type DayPart = 'morning' | 'afternoon' | 'evening' | 'night';

/** Morning 5–11, afternoon 12–17, evening 18–22, otherwise night (local hours, #189). */
export function dayPart(now: Date): DayPart {
  const h = now.getHours();
  if (h >= 5 && h <= 11) return 'morning';
  if (h >= 12 && h <= 17) return 'afternoon';
  if (h >= 18 && h <= 22) return 'evening';
  return 'night';
}

/** Message key of the greeting; with or without a name. */
export function greetingKey(part: DayPart, named: boolean): `appearance.greeting_${DayPart}${'' | '_named'}` {
  return named ? `appearance.greeting_${part}_named` : `appearance.greeting_${part}`;
}

/** Message keys of the line after the greeting (#512). */
export type MomentKey =
  | 'appearance.moment_night'
  | 'appearance.moment_weekend'
  | 'appearance.moment_weekend_evening'
  | 'appearance.moment_friday'
  | 'appearance.moment_monday'
  | 'appearance.moment_late'
  | 'appearance.moment_morning'
  | 'appearance.moment_day'
  | 'appearance.moment_evening';

/**
 * A short line after the greeting, from the local weekday and hour only
 * (#512): nothing is asked and nothing leaves the browser. Messages may use
 * {weekday}.
 */
export function momentKey(now: Date): MomentKey {
  const h = now.getHours();
  const day = now.getDay();
  const weekend = day === 0 || day === 6;
  if (h < 5) return 'appearance.moment_night';
  if (weekend) return h < 18 ? 'appearance.moment_weekend' : 'appearance.moment_weekend_evening';
  if (day === 5 && h >= 12) return 'appearance.moment_friday';
  if (day === 1 && h < 12) return 'appearance.moment_monday';
  if (h >= 22) return 'appearance.moment_late';
  if (h < 12) return 'appearance.moment_morning';
  if (h < 18) return 'appearance.moment_day';
  return 'appearance.moment_evening';
}
