/** Days the Tonight card forecasts, today included. */
export const TONIGHT_DAYS = 4;

/** What the card reads from a widget's configuration; the location is the weather source's business. */
export interface TonightConfig {
  /** IANA time zone, empty for the browser's zone. */
  timeZone: string;
  hour12: boolean;
}

/** Reads a stored configuration; the keys match those of the clock widget. */
export function parseTonightConfig(raw: Record<string, unknown>): TonightConfig {
  const zone = typeof raw['timeZone'] === 'string' ? raw['timeZone'] : '';
  return { timeZone: isValidTimeZone(zone) ? zone : '', hour12: raw['hour12'] === true };
}

function isValidTimeZone(tz: string): boolean {
  if (tz === '') return true;
  try {
    new Intl.DateTimeFormat('en', { timeZone: tz });
    return true;
  } catch {
    return false;
  }
}

/** One day of the outlook, already described in words and an icon kind. */
export interface TonightDay {
  date: Date;
  icon: string;
  label: string;
  max: number;
  min: number;
  /** Probability of precipitation in percent. */
  rain: number;
}

/** The current weather and the next days, as the weather source describes them. */
export interface TonightWeather {
  temperature: number;
  /** Felt temperature; absent when the source does not report it. */
  feelsLike?: number;
  condition: string;
  icon: string;
  /** Name of the place; empty when none was given. */
  place: string;
  days: TonightDay[];
}

export type WarningSeverity = 'minor' | 'moderate' | 'severe' | 'extreme';

/** An official weather warning. */
export interface TonightWarning {
  id: string;
  title: string;
  /** Validity such as "until Thu 17:00"; may be empty. */
  validity: string;
  severity: WarningSeverity;
}

const RANK: Record<WarningSeverity, number> = { minor: 0, moderate: 1, severe: 2, extreme: 3 };

/** The most severe warning, which the card shows; the first one wins a tie. */
export function worstWarning(warnings: readonly TonightWarning[]): TonightWarning | undefined {
  return warnings.reduce<TonightWarning | undefined>(
    (worst, w) => (!worst || RANK[w.severity] > RANK[worst.severity] ? w : worst),
    undefined,
  );
}

/**
 * A warning title in sentence case: weather services send event names such
 * as "STARKWIND" in capitals, which the chip of the card does not shout.
 * Titles with lower-case letters already are written as meant.
 */
export function sentenceCase(title: string): string {
  if (title !== title.toUpperCase()) return title;
  const lower = title.toLowerCase();
  return lower.charAt(0).toUpperCase() + lower.slice(1);
}

/** Severe and extreme warnings are problems; the others ask for attention. */
export function warningLevel(w: Pick<TonightWarning, 'severity'>): 'attention' | 'problem' {
  return RANK[w.severity] >= RANK.severe ? 'problem' : 'attention';
}

/** Whole degrees with the degree sign. */
export function degrees(v: number): string {
  return `${Math.round(v)}°`;
}

/** The time of day in the configured zone. */
export function formatTime(date: Date, cfg: TonightConfig, locale?: string): string {
  return new Intl.DateTimeFormat(locale, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: cfg.hour12,
    timeZone: cfg.timeZone || undefined,
  }).format(date);
}

/** Weekday, day and month in the configured zone. */
export function formatDate(date: Date, cfg: TonightConfig, locale?: string): string {
  return new Intl.DateTimeFormat(locale, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    timeZone: cfg.timeZone || undefined,
  }).format(date);
}
