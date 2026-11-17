import type { CountdownDoc } from '../../features/countdown/adapter/local/localCountdownGateway';
import type { FocusDoc } from '../../features/focus/adapter/local/localFocusGateway';
import type { HabitsDoc } from '../../features/habits/adapter/local/localHabitsGateway';
import { describe, expect, it } from 'vitest';
import { defaultAppearance } from '../../features/appearance/domain/appearance';
import type { DashboardDoc } from '../../features/dashboard/adapter/local/localDashboardGateway';
import { parseExportFile } from '../../features/transfer/domain/transfer';
import type { BookmarksDoc } from '../../features/bookmarks/adapter/local/localBookmarksGateway';
import type { NotesDoc } from '../../features/notes/adapter/local/localNotesGateway';
import type { TodosDoc } from '../../features/todos/adapter/local/localTodosGateway';
import { MemoryDocumentStore } from '../../shared/local/documentStore';
import { exportFileName, exportLocalData } from './localExport';

const now = new Date('2026-11-16T07:30:05Z');

/** Local data with every column the export writes. */
async function sampleStore(): Promise<MemoryDocumentStore> {
  const store = new MemoryDocumentStore();
  const doc: DashboardDoc = {
    tabs: [
      {
        id: '11111111-1111-4111-8111-111111111111',
        name: 'Home',
        position: 0,
        background: '',
        rules: [{ days: [1, 2, 3, 4, 5], from: '07:00', to: '09:30', networks: [] }],
        mobile: [],
        mobileCustom: false,
      },
      {
        id: '22222222-2222-4222-8222-222222222222',
        name: 'Evening',
        position: 1,
        background: 'linear-gradient(135deg, #1e3a8a, #9333ea)',
        rules: [{ days: [], from: '', to: '', networks: [] }],
        mobile: [{ id: '44444444-4444-4444-8444-444444444444', hidden: true, rows: 2 }],
        mobileCustom: true,
      },
    ],
    widgets: [
      {
        id: '33333333-3333-4333-8333-333333333333',
        tabId: '11111111-1111-4111-8111-111111111111',
        kind: 'clock',
        x: 0,
        y: 0,
        w: 3,
        h: 3,
        config: {},
      },
      {
        id: '44444444-4444-4444-8444-444444444444',
        tabId: '22222222-2222-4222-8222-222222222222',
        kind: 'notes',
        x: 0,
        y: 0,
        w: 4,
        h: 6,
        config: { scope: 'pinned' },
      },
    ],
    templates: [
      {
        id: '55555555-5555-4555-8555-555555555555',
        name: 'Start',
        widgets: [{ kind: 'clock', rect: { x: 0, y: 0, w: 3, h: 3 }, config: {} }],
      },
    ],
  };
  await store.put('dashboard', doc);
  await store.put('appearance', {
    ...defaultAppearance,
    theme: 'dark',
    greetingName: 'Sam',
    profile: 'everyday',
  });
  const notes: NotesDoc = {
    notes: [
      {
        id: '66666666-6666-4666-8666-666666666666',
        title: 'Shopping',
        body: 'milk, bread',
        pinned: true,
        createdAt: '2026-11-15T09:00:00.000Z',
        updatedAt: '2026-11-15T10:00:00.000Z',
        version: 2,
      },
    ],
    revisions: [
      {
        id: '77777777-7777-4777-8777-777777777777',
        noteId: '66666666-6666-4666-8666-666666666666',
        title: 'Shopping',
        body: 'milk',
        createdAt: '2026-11-15T09:00:00.000Z',
      },
    ],
  };
  await store.put('notes', notes);
  const todos: TodosDoc = {
    todos: [
      {
        id: '88888888-8888-4888-8888-888888888888',
        title: 'Water plants',
        due: '2026-11-18',
        position: 0,
        createdAt: '2026-11-15T09:00:00.000Z',
        completedAt: '',
        repeat: { kind: 'weekly', every: 1, days: [1, 4] },
        reminder: { at: '08:00', daysBefore: 0 },
        version: 1,
      },
      {
        id: '99999999-9999-4999-8999-999999999999',
        title: 'Call mum',
        due: '',
        position: 1,
        createdAt: '2026-11-15T09:05:00.000Z',
        completedAt: '2026-11-15T18:00:00.000Z',
        repeat: { kind: '', every: 1, days: [] },
        reminder: { at: '', daysBefore: 0 },
        version: 2,
      },
    ],
  };
  await store.put('todos', todos);
  const bookmarks: BookmarksDoc = {
    folders: [{ id: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', name: 'Work', position: 0 }],
    bookmarks: [
      {
        id: 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
        folderId: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
        title: 'Docs',
        url: 'https://example.org/docs',
        description: 'Handbook',
        tags: ['docs', 'work'],
        createdAt: '2026-11-15T09:00:00.000Z',
      },
      {
        id: 'cccccccc-cccc-4ccc-8ccc-cccccccccccc',
        folderId: '',
        title: 'News',
        url: 'https://news.example',
        description: '',
        tags: [],
        createdAt: '2026-11-15T09:10:00.000Z',
      },
    ],
  };
  await store.put('bookmarks', bookmarks);
  const countdown: CountdownDoc = {
    items: [
      {
        id: '66666666-6666-4666-8666-666666666666',
        title: 'Holiday',
        emoji: '🏖️',
        date: '2026-12-20',
        time: '',
        createdAt: '2026-11-15T10:00:00.000Z',
      },
      {
        id: '77777777-7777-4777-8777-777777777777',
        title: 'Dentist',
        emoji: '',
        date: '2026-11-18',
        time: '14:30',
        createdAt: '2026-11-15T10:05:00.000Z',
      },
    ],
  };
  await store.put('countdown', countdown);
  const focus: FocusDoc = {
    sessions: [
      { id: '88888888-8888-4888-8888-888888888888', minutes: 25, endedAt: '2026-11-16T07:00:00.000Z' },
    ],
  };
  await store.put('focus', focus);
  const habits: HabitsDoc = {
    habits: [
      {
        id: '99999999-9999-4999-8999-999999999999',
        name: 'Read',
        emoji: '📚',
        days: [5, 1, 3],
        position: 0,
        createdAt: '2026-11-01T08:00:00.000Z',
      },
    ],
    completions: [
      { habitId: '99999999-9999-4999-8999-999999999999', day: '2026-11-13' },
      { habitId: '99999999-9999-4999-8999-999999999999', day: '2026-11-16' },
      // Left over from a deleted habit: not exported.
      { habitId: 'deleted', day: '2026-11-16' },
    ],
  };
  await store.put('habits', habits);
  return store;
}

describe('exportLocalData', () => {
  it('writes the server export format; the file is imported by a Go test (TestLocalModeExportImports)', async () => {
    const file = await exportLocalData(await sampleStore(), now);
    expect(typeof parseExportFile(JSON.stringify(file))).not.toBe('string');
    await expect(JSON.stringify(file, null, 2) + '\n').toMatchFileSnapshot('./testdata/local-export.json');
  });

  it('leaves out features without stored data, so the import keeps the server data for them', async () => {
    const file = await exportLocalData(new MemoryDocumentStore(), now);
    expect(file.sections).toEqual({});
  });

  it('names the file like the server does', () => {
    expect(exportFileName(new Date(2026, 10, 16, 7, 30, 5))).toBe('waylight-export-20261116-073005.json');
  });
});
