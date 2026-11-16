import { EXPORT_OWNER, exportDay, exportTime, type ExportTables } from '../../../../shared/local/exportRows';
import type { TodosDoc } from './localTodosGateway';

/**
 * The to-dos document as rows of the todos table (migrations of
 * internal/modules/todos). Only the built-in list exists locally, so no
 * todos_sources rows are written.
 */
export function todosExport(doc: TodosDoc): ExportTables {
  return {
    todos: doc.todos.map((t) => ({
      id: t.id,
      title: t.title,
      due: t.due ? exportDay(t.due) : null,
      position: t.position,
      created_at: exportTime(t.createdAt),
      completed_at: t.completedAt ? exportTime(t.completedAt) : null,
      repeat_kind: t.repeat.kind,
      repeat_every: t.repeat.kind === '' ? 0 : t.repeat.every,
      // The server stores weekdays as "1,4" (formatDays in its gateway).
      repeat_days: t.repeat.days.join(','),
      remind_at: t.reminder.at,
      remind_days_before: t.reminder.daysBefore,
      reminded_for: null,
      version: t.version,
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
  };
}
