import { EXPORT_OWNER, exportTime, type ExportTables } from '../../../../shared/local/exportRows';
import type { HabitsDoc } from './localHabitsGateway';

/**
 * The habits and their completed days as rows of habits and
 * habit_completions (migrations of internal/modules/habits). Weekdays are
 * stored as "1,3,5", empty for every day.
 */
export function habitsExport(doc: HabitsDoc): ExportTables {
  const ids = new Set(doc.habits.map((h) => h.id));
  return {
    habits: doc.habits.map((h) => ({
      id: h.id,
      name: h.name,
      emoji: h.emoji,
      days: [...h.days].sort((a, b) => a - b).join(','),
      position: h.position,
      created_at: exportTime(h.createdAt),
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
    // Completions of a deleted habit would break the foreign key on import.
    habit_completions: doc.completions
      .filter((c) => ids.has(c.habitId))
      .map((c) => ({ habit_id: c.habitId, day: c.day })),
  };
}
