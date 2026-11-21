import { describe, expect, it } from 'vitest';
import { loadAreaStatuses } from './settingsStatus';
import { loadRecent, rememberChange } from './settingsRecent';

const now = new Date(2026, 10, 21, 12, 0);

describe('area statuses on the settings start page (#600)', () => {
  it('reads each status from its source', async () => {
    const s = await loadAreaStatuses(
      ['account', 'notifications', 'connections', 'data', 'system'],
      {
        twoFactor: async () => ({ enabled: false }),
        channels: async () => [{ enabled: true }, { enabled: false }, { enabled: true }],
        backups: async () => [
          { createdAt: new Date(2026, 10, 20, 3, 0) },
          { createdAt: new Date(2026, 10, 21, 3, 0) },
        ],
        version: async () => ({ current: '0.2.0', updateAvailable: true, latest: { version: '0.3.0' } }),
        connections: 4,
      },
      now,
    );
    expect(s.account).toEqual({ text: 'Two-factor off', tone: 'warn' });
    expect(s.notifications).toEqual({ text: '2 channels on', tone: 'ok' });
    expect(s.connections).toEqual({ text: '4 integrations', tone: 'info' });
    expect(s.data?.tone).toBe('ok');
    expect(s.data?.text).toMatch(/^Backup today 03:00/);
    expect(s.system).toEqual({ text: 'Update to 0.3.0 available', tone: 'warn' });
  });

  it('warns about a missing or stale backup', async () => {
    const none = await loadAreaStatuses(['data'], { backups: async () => [] }, now);
    expect(none.data).toEqual({ text: 'No backup yet', tone: 'warn' });
    const old = await loadAreaStatuses(
      ['data'],
      { backups: async () => [{ createdAt: new Date(2026, 10, 10, 3, 0) }] },
      now,
    );
    expect(old.data?.tone).toBe('warn');
    expect(old.data?.text).toMatch(/^Last backup/);
  });

  it('leaves a tile without status when its source is missing or fails', async () => {
    const s = await loadAreaStatuses(
      ['account', 'system', 'people'],
      {
        twoFactor: async () => {
          throw new Error('403');
        },
      },
      now,
    );
    expect(s).toEqual({});
  });
});

describe('recently changed sections (#600)', () => {
  const memory = () => {
    const data = new Map<string, string>();
    return {
      getItem: (k: string) => data.get(k) ?? null,
      setItem: (k: string, v: string) => void data.set(k, v),
    };
  };

  it('remembers the last three changes in this browser', () => {
    const storage = memory();
    for (const id of ['a', 'b', 'c', 'd', 'b']) rememberChange(id, storage);
    expect(loadRecent(storage)).toEqual(['b', 'd', 'c']);
  });

  it('copes with broken or blocked storage', () => {
    expect(loadRecent({ getItem: () => '{oops' })).toEqual([]);
    expect(loadRecent({ getItem: () => '[1,"a"]' })).toEqual(['a']);
    const blocked = {
      getItem: () => {
        throw new Error('SecurityError');
      },
      setItem: () => {
        throw new Error('SecurityError');
      },
    };
    expect(rememberChange('a', blocked)).toEqual(['a']);
  });
});
