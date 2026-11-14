import { t } from '../../../shared/i18n';

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

/** The hours grouped by calendar day (today, tomorrow), in order. */
export function byDay(hours: readonly PriceHour[]): { day: string; hours: PriceHour[] }[] {
  const out: { day: string; hours: PriceHour[] }[] = [];
  for (const h of hours) {
    const day = `${h.start.getFullYear()}-${h.start.getMonth()}-${h.start.getDate()}`;
    const last = out.at(-1);
    if (last?.day === day) last.hours.push(h);
    else out.push({ day, hours: [h] });
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
  const values = hours.map((h) => h.centPerKWh);
  const max = Math.max(0, ...values);
  const min = Math.min(0, ...values);
  const span = max - min || 1;
  return {
    heights: values.map((v) => (Math.abs(v) / span) * 100),
    zero: (max / span) * 100,
  };
}

/** "12.34 ct" in the UI language. */
export const formatCent = (v: number, locale: string): string =>
  `${v.toLocaleString(locale, { minimumFractionDigits: 1, maximumFractionDigits: 2 })} ct`;
