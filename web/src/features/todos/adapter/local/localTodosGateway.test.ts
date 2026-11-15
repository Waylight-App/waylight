import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import type { Repeat } from '../../domain/todo';
import { TodosUseCases } from '../../usecase/todosUseCases';
import { LocalTodosGateway, nextDue } from './localTodosGateway';

const rule = (kind: Repeat['kind'], every: number, ...days: number[]): Repeat => ({ kind, every, days });

describe('nextDue', () => {
  // The same cases as the server's TestRecurrenceNext; 2026-11-12 is a Thursday.
  it.each([
    ['daily on time', rule('daily', 1), '2026-11-12', '2026-11-12', '2026-11-13'],
    ['daily done early', rule('daily', 1), '2026-11-12', '2026-11-10', '2026-11-13'],
    ['overdue daily skips past days', rule('daily', 1), '2026-11-09', '2026-11-12', '2026-11-13'],
    ['every 3 days', rule('daily', 3), '2026-11-12', '2026-11-12', '2026-11-15'],
    ['weekly Mon+Thu', rule('weekly', 1, 1, 4), '2026-11-12', '2026-11-12', '2026-11-16'],
    ['weekly Mon+Thu from Monday', rule('weekly', 1, 1, 4), '2026-11-16', '2026-11-16', '2026-11-19'],
    ['every 2 weeks on Thursday', rule('weekly', 2, 4), '2026-11-12', '2026-11-12', '2026-11-26'],
    ['monthly', rule('monthly', 1), '2026-11-12', '2026-11-12', '2026-12-12'],
    ['monthly from the 31st clamps', rule('monthly', 1), '2027-01-31', '2027-01-31', '2027-02-28'],
    [
      'monthly keeps the 31st after a short month',
      rule('monthly', 1),
      '2027-01-31',
      '2027-03-01',
      '2027-03-31',
    ],
    ['quarterly', rule('monthly', 3), '2026-11-12', '2026-11-12', '2027-02-12'],
    ['after completion', rule('after', 10), '2026-11-01', '2026-11-12', '2026-11-22'],
    ['no due date counts from completion', rule('daily', 2), '', '2026-11-12', '2026-11-14'],
  ])('%s', (_name, repeat, due, doneOn, want) => {
    expect(nextDue(repeat, due, doneOn)).toBe(want);
  });
});

describe('LocalTodosGateway', () => {
  const now = () => new Date('2026-11-12T10:00:00');

  it('adds, completes and clears to-dos without a server', async () => {
    const uc = new TodosUseCases(new LocalTodosGateway(new MemoryDocumentStore(), now));
    const a = await uc.add('Buy milk @2026-11-10');
    await uc.add('Call mum');
    expect(a.due).toBe('2026-11-10');
    expect(a.overdue).toBe(true);
    await uc.toggle(a);
    expect((await uc.list(false)).map((t) => t.title)).toEqual(['Call mum']);
    expect(await uc.list(true)).toHaveLength(2);
    expect(await new LocalTodosGateway(new MemoryDocumentStore(), now).clearCompleted()).toBe(0);
  });

  it('creates the next occurrence when a recurring to-do is done', async () => {
    const gw = new LocalTodosGateway(new MemoryDocumentStore(), now);
    const t = await gw.create('Water plants', '2026-11-12');
    await gw.update(t.id, t.title, t.due, false, {
      repeat: rule('daily', 2),
      reminder: { at: '', daysBefore: 0 },
    });
    await gw.update(t.id, t.title, t.due, true);
    const open = await gw.list(false);
    expect(open).toHaveLength(1);
    expect(open[0]?.due).toBe('2026-11-14');
    expect(open[0]?.repeat.kind).toBe('daily');
    expect(await gw.clearCompleted()).toBe(1);
  });
});
