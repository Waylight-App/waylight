import { describe, expect, it } from 'vitest';
import { defaultAppearance } from '../../features/appearance/domain/appearance';
import type { DashboardDoc } from '../../features/dashboard/adapter/local/localDashboardGateway';
import { parseExportFile } from '../../features/transfer/domain/transfer';
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
    expect(exportFileName(new Date(2026, 10, 16, 7, 30, 5))).toBe('trisse-home-export-20261116-073005.json');
  });
});
