/**
 * Rows of the server's export format (internal/shared/sqlutil.Rows): one
 * object per table row, column name → value. The extension's local mode
 * writes them so that a file it exports imports into a server through the
 * existing transfer import (ADR 0019). Only strings, numbers, booleans and
 * null are allowed; JSON columns are stored as strings, as in SQLite.
 */
export type ExportValue = string | number | boolean | null;
export type ExportRow = Record<string, ExportValue>;
/** The tables of one module's section, parents before children. */
export type ExportTables = Record<string, ExportRow[]>;

/** Account that owns everything exported from local mode: the server's owner (ADR 0018). */
export const EXPORT_OWNER = 'owner';

/**
 * A time as the server stores it (internal/shared/sqlutil.FormatTime): UTC
 * with nine fractional digits. The server's parser expects exactly this
 * layout, so an ISO string from the browser has to be widened.
 */
export function exportTime(at: Date | string): string {
  const iso = (typeof at === 'string' ? new Date(at) : at).toISOString(); // …T07:30:05.123Z
  return `${iso.slice(0, 23)}000000Z`;
}

/** A calendar day ("YYYY-MM-DD") as the server stores due dates: midnight UTC. */
export const exportDay = (day: string): string => `${day}T00:00:00.000000000Z`;
