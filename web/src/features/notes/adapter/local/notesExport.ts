import { EXPORT_OWNER, exportTime, type ExportTables } from '../../../../shared/local/exportRows';
import type { NotesDoc } from './localNotesGateway';

/**
 * The notes document as rows of notes and note_revisions (migrations of
 * internal/modules/notes). Notes belong to the owner and stay private.
 */
export function notesExport(doc: NotesDoc): ExportTables {
  return {
    notes: doc.notes.map((n) => ({
      id: n.id,
      title: n.title,
      body: n.body,
      pinned: n.pinned ? 1 : 0,
      created_at: exportTime(n.createdAt),
      updated_at: exportTime(n.updatedAt),
      version: n.version,
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
    note_revisions: doc.revisions.map((r) => ({
      id: r.id,
      note_id: r.noteId,
      title: r.title,
      body: r.body,
      created_at: exportTime(r.createdAt),
    })),
  };
}
