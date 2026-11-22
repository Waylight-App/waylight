import { t } from '../../../shared/i18n';
import { barHeights } from '../../../shared/ui/bars';

/** Bidding zones (#288). */
export const ZONES = ['DE-LU', 'AT'] as const;
export type Zone = (typeof ZONES)[number];

/** One hour of the day-ahead market. */
export interface PriceHour {
  start: Date;
  centPerKWh: number;
  cheapest: boolean;
}

export interface ElectricityPrices {
  zone: Zone;
  hours: PriceHour[];
  fetchedAt: Date;
  stale: boolean;
}

/** Widget configuration: the bidding zone. */
export function parseConfig(raw: Record<string, unknown>): { zone: Zone } {
  const z = raw['zone'];
  return { zone: ZONES.includes(z as Zone) ? (z as Zone) : 'DE-LU' };
}

/** Name of a zone for the settings. */
export const zoneName = (z: Zone): string => t(z === 'AT' ? 'electricity.zone_at' : 'electricity.zone_de');

const dayKey = (d: Date): string => `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`;

/** A calendar day of prices, relative to now. */
export interface PriceDay {
  day: string;
  when: 'today' | 'tomorrow';
  hours: PriceHour[];
}

/**
 * The hours of today and tomorrow, grouped by local calendar day. Days are
 * labelled by their date, not their position, so prices that are still from
 * yesterday (a stale answer shortly after midnight) never show as "today".
 */
export function byDay(hours: readonly PriceHour[], now: Date): PriceDay[] {
  const today = dayKey(now);
  const tomorrow = dayKey(new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1));
  const out: PriceDay[] = [];
  for (const h of hours) {
    const day = dayKey(h.start);
    const when = day === today ? 'today' : day === tomorrow ? 'tomorrow' : null;
    if (!when) continue;
    const last = out.at(-1);
    if (last?.day === day) last.hours.push(h);
    else out.push({ day, when, hours: [h] });
  }
  return out;
}

/** The hour that contains now, if any. */
export function currentHour(hours: readonly PriceHour[], now: Date): PriceHour | undefined {
  return hours.find(
    (h) => h.start.getTime() <= now.getTime() && now.getTime() < h.start.getTime() + 3_600_000,
  );
}

/**
 * Bar heights in percent of the chart: the scale includes zero, so negative
 * prices hang below the base line. Returns the base line position too.
 */
export function bars(hours: readonly PriceHour[]): { heights: number[]; zero: number } {
  return barHeights(hours.map((h) => h.centPerKWh));
}

/** "12.34 ct" in the UI language. */
export const formatCent = (v: number, locale: string): string =>
  `${v.toLocaleString(locale, { minimumFractionDigits: 1, maximumFractionDigits: 2 })} ct`;
