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
