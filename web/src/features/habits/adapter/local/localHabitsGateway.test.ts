import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import { LocalHabitsGateway, stats, type HabitRow } from './localHabitsGateway';

const habit = (days: number[] = []): HabitRow => ({
  id: 'h',
  name: 'Read',
  emoji: '📚',
  days,
  position: 0,
  createdAt: new Date('2026-11-01T08:00:00').toISOString(),
});

describe('stats (mirrors ComputeStats)', () => {
  it('counts the current streak and keeps it while today is open', () => {
    const s = stats(habit(), new Set(['2026-11-14', '2026-11-15']), '2026-11-16');
    expect(s).toMatchObject({ doneToday: false, current: 2, best: 2 });
    expect(stats(habit(), new Set(['2026-11-14', '2026-11-15', '2026-11-16']), '2026-11-16').current).toBe(3);
  });

  it('skips days the habit is not due', () => {
    // Monday and Wednesday only; 2026-11-16 is a Monday.
    const s = stats(habit([1, 3]), new Set(['2026-11-11', '2026-11-16']), '2026-11-16');
    expect(s.current).toBe(2);
  });

  it('keeps the best streak after a break', () => {
    const s = stats(habit(), new Set(['2026-11-02', '2026-11-03', '2026-11-04', '2026-11-15']), '2026-11-16');
    expect(s).toMatchObject({ current: 1, best: 3 });
  });

  it('has a 28-day heatmap ending today', () => {
    const s = stats(habit([1]), new Set(['2026-11-16']), '2026-11-16');
    expect(s.heatmap).toHaveLength(28);
    expect(s.heatmap.at(-1)).toEqual({ date: '2026-11-16', due: true, done: true });
    expect(s.heatmap.at(-2)?.due).toBe(false);
  });
});

describe('LocalHabitsGateway', () => {
  const now = () => new Date('2026-11-16T12:00:00');

  it('checks off today and yesterday, nothing older', async () => {
    const g = new LocalHabitsGateway(new MemoryDocumentStore(), now);
    const h = await g.create({ name: 'Walk', emoji: '', days: [] });
    expect((await g.toggle(h.id, '2026-11-16')).doneToday).toBe(true);
    expect((await g.toggle(h.id, '2026-11-15')).current).toBe(2);
    expect((await g.toggle(h.id, '2026-11-16')).doneToday).toBe(false);
    await expect(g.toggle(h.id, '2026-11-10')).rejects.toThrow();
  });

  it('lists in order of creation and removes completions with the habit', async () => {
    const store = new MemoryDocumentStore();
    const g = new LocalHabitsGateway(store, now);
    const a = await g.create({ name: 'A', emoji: '', days: [] });
    await g.create({ name: 'B', emoji: '', days: [] });
    await g.toggle(a.id, '2026-11-16');
    expect((await g.list()).map((h) => h.name)).toEqual(['A', 'B']);
    await g.remove(a.id);
    expect((await g.list()).map((h) => h.name)).toEqual(['B']);
    expect((await store.get<{ completions: unknown[] }>('habits'))?.completions).toEqual([]);
  });
});
