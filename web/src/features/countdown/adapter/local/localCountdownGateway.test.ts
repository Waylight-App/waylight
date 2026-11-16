import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import { hiddenAt, LocalCountdownGateway } from './localCountdownGateway';

const at = (s: string) => new Date(s);

describe('hiddenAt (mirrors Countdown.Hidden)', () => {
  const allDay = { id: '1', title: 'Exam', emoji: '', date: '2026-11-16', time: '' };
  it('keeps an all-day countdown running until the end of its day', () => {
    expect(hiddenAt(allDay, at('2026-11-16T23:59:00'), 0)).toBe(false);
    expect(hiddenAt(allDay, at('2026-11-17T00:00:00'), 0)).toBe(true);
  });
  it('shows a past countdown for the grace days', () => {
    expect(hiddenAt(allDay, at('2026-11-17T12:00:00'), 1)).toBe(false);
    expect(hiddenAt(allDay, at('2026-11-18T00:00:00'), 1)).toBe(true);
  });
  it('ends a timed countdown at its moment', () => {
    const timed = { ...allDay, time: '14:30' };
    expect(hiddenAt(timed, at('2026-11-16T14:29:00'), 0)).toBe(false);
    expect(hiddenAt(timed, at('2026-11-16T14:30:00'), 0)).toBe(true);
  });
});

describe('LocalCountdownGateway', () => {
  it('lists by moment, then title, and hides old ones', async () => {
    const g = new LocalCountdownGateway(new MemoryDocumentStore(), () => at('2026-11-16T12:00:00'));
    await g.create({ title: 'Trip', emoji: '✈️', date: '2026-12-01', time: '' });
    await g.create({ title: 'B', emoji: '', date: '2026-11-20', time: '' });
    await g.create({ title: 'A', emoji: '', date: '2026-11-20', time: '' });
    await g.create({ title: 'Old', emoji: '', date: '2026-11-01', time: '' });
    expect((await g.list(1)).map((c) => c.title)).toEqual(['A', 'B', 'Trip']);
  });

  it('updates and removes', async () => {
    const g = new LocalCountdownGateway(new MemoryDocumentStore(), () => at('2026-11-16T12:00:00'));
    const c = await g.create({ title: 'X', emoji: '', date: '2026-11-20', time: '' });
    await g.update(c.id, { title: 'Y', emoji: '🎉', date: '2026-11-21', time: '18:00' });
    expect(await g.list(0)).toEqual([
      { id: c.id, title: 'Y', emoji: '🎉', date: '2026-11-21', time: '18:00' },
    ]);
    await g.remove(c.id);
    expect(await g.list(0)).toEqual([]);
  });
});
