import { describe, expect, it } from 'vitest';
import { type BackupStatus, dayWord, downloadUrl, healthLine, restoreForm } from './backup';

describe('downloadUrl', () => {
  it('encodes the file name', () => {
    expect(downloadUrl({ name: 'waylight-20261111-030000.db', size: 1, createdAt: new Date() })).toBe(
      '/api/backups/waylight-20261111-030000.db',
    );
  });
});

describe('healthLine (#677)', () => {
  const base: BackupStatus = {
    health: 'ok',
    lastBackupAt: new Date(2026, 10, 21, 3, 0),
    count: 3,
    failedAt: null,
    error: '',
    offsiteFailedAt: null,
    offsiteError: '',
  };

  it('puts a failure before everything else', () => {
    const failedAt = new Date(2026, 10, 21, 4, 0);
    expect(healthLine({ ...base, health: 'failed', failedAt, offsiteFailedAt: failedAt })).toEqual({
      tone: 'error',
      key: 'backup.health.failed',
      at: failedAt,
    });
  });

  it('tells a missing backup, a failed off-site copy and an old backup apart', () => {
    expect(healthLine({ ...base, health: 'stale', lastBackupAt: null }).key).toBe('backup.health.none');
    expect(healthLine({ ...base, health: 'stale', offsiteFailedAt: new Date() }).key).toBe(
      'backup.health.offsite_failed',
    );
    expect(healthLine({ ...base, health: 'stale' })).toMatchObject({
      tone: 'warn',
      key: 'backup.health.stale',
    });
    expect(healthLine(base)).toEqual({ tone: 'ok', key: 'backup.health.ok', at: base.lastBackupAt });
  });
});

describe('dayWord', () => {
  const now = new Date(2026, 10, 21, 12, 0);
  it('names today and yesterday', () => {
    expect(dayWord(new Date(2026, 10, 21, 3, 0), now)).toBe('today');
    expect(dayWord(new Date(2026, 10, 20, 23, 0), now)).toBe('yesterday');
    expect(dayWord(new Date(2026, 10, 19, 23, 0), now)).toBe('earlier');
  });
});

describe('restoreForm', () => {
  const base = {
    password: 'pw',
    code: ' 123 456 ',
    source: 'local' as const,
    name: 'waylight-20261111-030000.db',
  };

  it('sends the confirmation word only when the owner confirmed', () => {
    expect(restoreForm({ ...base, confirmed: true }).get('confirm')).toBe('replace');
    expect(restoreForm({ ...base, confirmed: false }).get('confirm')).toBeNull();
  });

  it('puts the file last, after every text field', () => {
    const file = new File(['x'], 'backup.db');
    const keys = [...restoreForm({ ...base, source: 'upload', confirmed: true, file }).keys()];
    expect(keys.at(-1)).toBe('file');
    expect(restoreForm({ ...base, confirmed: true }).get('code')).toBe('123 456');
  });
});
