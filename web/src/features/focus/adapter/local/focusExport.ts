import { EXPORT_OWNER, exportTime, type ExportTables } from '../../../../shared/local/exportRows';
import type { FocusDoc } from './localFocusGateway';

/** The finished sessions as rows of focus_sessions (migrations of internal/modules/focus). */
export function focusExport(doc: FocusDoc): ExportTables {
  return {
    focus_sessions: doc.sessions.map((s) => ({
      id: s.id,
      minutes: s.minutes,
      ended_at: exportTime(s.endedAt),
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
  };
}
