import type { Severity } from './notifications';

/** One notification shown live in the web interface (#583). */
export interface LiveNotification {
  title: string;
  body: string;
  severity: Severity;
  type: string;
  /** ISO time the server sent it. */
  at: string;
}

const severities: readonly string[] = ['info', 'warning', 'critical'];

/** Reads the data of a stream event; anything else gives null. */
export function parseLiveNotification(data: string): LiveNotification | null {
  let v: unknown;
  try {
    v = JSON.parse(data);
  } catch {
    return null;
  }
  if (typeof v !== 'object' || v === null) return null;
  const o = v as Record<string, unknown>;
  if (typeof o.title !== 'string' || typeof o.body !== 'string') return null;
  return {
    title: o.title,
    body: o.body,
    severity: (severities.includes(o.severity as string) ? o.severity : 'info') as Severity,
    type: typeof o.type === 'string' ? o.type : '',
    at: typeof o.at === 'string' ? o.at : '',
  };
}
