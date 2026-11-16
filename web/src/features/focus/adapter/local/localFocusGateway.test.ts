import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import { LocalFocusGateway, week } from './localFocusGateway';

describe('week (mirrors focus domain Week)', () => {
  it('counts sessions per day for the last seven days, oldest first', () => {
    const days = week(
      [
        { id: 'a', minutes: 25, endedAt: new Date('2026-11-16T09:00:00').toISOString() },
        { id: 'b', minutes: 25, endedAt: new Date('2026-11-16T10:00:00').toISOString() },
        { id: 'c', minutes: 50, endedAt: new Date('2026-11-10T10:00:00').toISOString() },
        { id: 'd', minutes: 25, endedAt: new Date('2026-11-09T10:00:00').toISOString() },
      ],
      new Date('2026-11-16T12:00:00'),
    );
    expect(days.map((d) => d.date)).toEqual([
      '2026-11-10',
      '2026-11-11',
      '2026-11-12',
      '2026-11-13',
      '2026-11-14',
      '2026-11-15',
      '2026-11-16',
    ]);
    expect(days[0]).toEqual({ date: '2026-11-10', sessions: 1, minutes: 50 });
    expect(days[6]).toEqual({ date: '2026-11-16', sessions: 2, minutes: 50 });
  });
});

describe('LocalFocusGateway', () => {
  it('records sessions and returns the week', async () => {
    const now = new Date('2026-11-16T12:00:00');
    const g = new LocalFocusGateway(new MemoryDocumentStore(), () => now);
    await g.record(25, new Date('2026-11-16T11:00:00'));
    const days = await g.record(25, new Date('2026-11-16T11:30:00'));
    expect(days.at(-1)).toEqual({ date: '2026-11-16', sessions: 2, minutes: 50 });
    expect((await g.stats()).at(-1)?.sessions).toBe(2);
  });
});
