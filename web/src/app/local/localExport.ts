/**
 * Export of the extension's local mode (ADR 0019, step 4): writes the local
 * data as a trisse-home export file, so moving to a server is the server's
 * existing import (Settings → Transfer). Each feature with a local gateway
 * contributes one section, named like its server module; a feature without
 * stored data is left out, so the import keeps the server's data for it.
 */
import { appearanceExport } from '../../features/appearance/adapter/local/appearanceExport';
import { LocalAppearanceGateway } from '../../features/appearance/adapter/local/localAppearanceGateway';
import { countdownExport } from '../../features/countdown/adapter/local/countdownExport';
import type { CountdownDoc } from '../../features/countdown/adapter/local/localCountdownGateway';
import { focusExport } from '../../features/focus/adapter/local/focusExport';
import type { FocusDoc } from '../../features/focus/adapter/local/localFocusGateway';
import { habitsExport } from '../../features/habits/adapter/local/habitsExport';
import type { HabitsDoc } from '../../features/habits/adapter/local/localHabitsGateway';
import { dashboardExport } from '../../features/dashboard/adapter/local/dashboardExport';
import type { DashboardDoc } from '../../features/dashboard/adapter/local/localDashboardGateway';
import { bookmarksExport } from '../../features/bookmarks/adapter/local/bookmarksExport';
import type { BookmarksDoc } from '../../features/bookmarks/adapter/local/localBookmarksGateway';
import type { NotesDoc } from '../../features/notes/adapter/local/localNotesGateway';
import { notesExport } from '../../features/notes/adapter/local/notesExport';
import type { TodosDoc } from '../../features/todos/adapter/local/localTodosGateway';
import { todosExport } from '../../features/todos/adapter/local/todosExport';
import { EXPORT_FORMAT, type ExportFile } from '../../features/transfer/domain/transfer';
import type { DocumentStore } from '../../shared/local/documentStore';
import type { ExportTables } from '../../shared/local/exportRows';

/** Version of the export format the server reads (internal/modules/transfer/domain). */
const EXPORT_VERSION = 1;

/** One module's part of the export. */
export interface LocalExportSection {
  /** Module name on the server, e.g. "dashboard". */
  name: string;
  /** The rows of the module's tables, or undefined when nothing is stored. */
  tables(store: DocumentStore): Promise<ExportTables | undefined>;
}

/** The sections of the features that have a local implementation. */
export const localExportSections: LocalExportSection[] = [
  {
    name: 'dashboard',
    async tables(store) {
      const doc = await store.get<DashboardDoc>('dashboard');
      return doc && dashboardExport(doc);
    },
  },
  {
    name: 'appearance',
    async tables(store) {
      if ((await store.get('appearance')) === undefined) return undefined;
      return appearanceExport(await new LocalAppearanceGateway(store).load());
    },
  },
  {
    name: 'notes',
    async tables(store) {
      const doc = await store.get<NotesDoc>('notes');
      return doc && notesExport(doc);
    },
  },
  {
    name: 'todos',
    async tables(store) {
      const doc = await store.get<TodosDoc>('todos');
      return doc && todosExport(doc);
    },
  },
  {
    name: 'bookmarks',
    async tables(store) {
      const doc = await store.get<BookmarksDoc>('bookmarks');
      return doc && bookmarksExport(doc);
    },
  },
  {
    name: 'countdown',
    async tables(store) {
      const doc = await store.get<CountdownDoc>('countdown');
      return doc && countdownExport(doc);
    },
  },
  {
    name: 'focus',
    async tables(store) {
      const doc = await store.get<FocusDoc>('focus');
      return doc && focusExport(doc);
    },
  },
  {
    name: 'habits',
    async tables(store) {
      const doc = await store.get<HabitsDoc>('habits');
      return doc && habitsExport(doc);
    },
  },
];

/** An export file with everything the local mode has stored. */
export async function exportLocalData(
  store: DocumentStore,
  now: Date = new Date(),
  sections: readonly LocalExportSection[] = localExportSections,
): Promise<ExportFile & { exportedAt: string }> {
  const out: Record<string, ExportTables> = {};
  for (const s of sections) {
    const tables = await s.tables(store);
    if (tables) out[s.name] = tables;
  }
  return { format: EXPORT_FORMAT, version: EXPORT_VERSION, exportedAt: now.toISOString(), sections: out };
}

/** File name in the server's pattern (trisse-home-export-YYYYMMDD-HHMMSS.json). */
export function exportFileName(now: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `trisse-home-export-${now.getFullYear()}${p(now.getMonth() + 1)}${p(now.getDate())}-${p(now.getHours())}${p(now.getMinutes())}${p(now.getSeconds())}.json`;
}
