import { newId } from '../../../../shared/id/uuid';
import type { DocumentStore } from '../../../../shared/local/documentStore';
import type { Habit, HabitInput } from '../../domain/habits';
import type { HabitsGateway } from '../../usecase/habitsUseCases';

/** Days of the heatmap, ending today (HeatmapDays on the server). */
const HEATMAP_DAYS = 28;

/** A habit as stored, in the shape of the server's habits table. */
export interface HabitRow {
  id: string;
  name: string;
  emoji: string;
  days: number[];
  position: number;
  /** ISO timestamp. */
  createdAt: string;
}

/** A completed day (habit_completions on the server). */
export interface CompletionRow {
  habitId: string;
  /** YYYY-MM-DD */
  day: string;
}

export interface HabitsDoc {
  habits: HabitRow[];
  completions: CompletionRow[];
}

const isoDay = (d: Date): string =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;

/** A YYYY-MM-DD day as a local date at noon, safe from DST shifts when stepping by days. */
const dayDate = (day: string): Date => new Date(`${day}T12:00:00`);
const addDays = (day: string, n: number): string => {
  const d = dayDate(day);
  d.setDate(d.getDate() + n);
  return isoDay(d);
};
const due = (h: HabitRow, day: string): boolean =>
  h.days.length === 0 || h.days.includes(dayDate(day).getDay());

/**
 * Streaks and heatmap of a habit. Mirrors ComputeStats in
 * internal/modules/habits/domain: a streak counts due days in a row that were
 * done; days the habit is not due neither count nor break it, and today does
 * not break the streak while it is still open.
 */
export function stats(h: HabitRow, doneDays: ReadonlySet<string>, today: string): Habit {
  const created = isoDay(new Date(h.createdAt));
  // The earliest completion, or the creation day; imports may hold older completions.
  const first = [...doneDays].reduce((min, d) => (d < min ? d : min), created);

  let current = 0;
  let day = due(h, today) && !doneDays.has(today) ? addDays(today, -1) : today;
  for (; day >= first; day = addDays(day, -1)) {
    if (!due(h, day)) continue;
    if (!doneDays.has(day)) break;
    current++;
  }

  let best = 0;
  let run = 0;
  for (let d = first; d <= today; d = addDays(d, 1)) {
    if (!due(h, d)) continue;
    if (doneDays.has(d)) best = Math.max(best, ++run);
    else if (d !== today) run = 0;
  }

  const heatmap = [];
  for (let i = HEATMAP_DAYS - 1; i >= 0; i--) {
    const d = addDays(today, -i);
    heatmap.push({ date: d, due: due(h, d), done: doneDays.has(d) });
  }
  return {
    id: h.id,
    name: h.name,
    emoji: h.emoji,
    days: h.days,
    doneToday: doneDays.has(today),
    current,
    best,
    heatmap,
  };
}

/**
 * Habits gateway of the browser extension's local mode (ADR 0019): habits and
 * completed days in one document, statistics computed on read as on the
 * server. Like the server, only today and yesterday can be checked off.
 */
export class LocalHabitsGateway implements HabitsGateway {
  constructor(
    private readonly store: DocumentStore,
    private readonly now: () => Date = () => new Date(),
  ) {}

  private async read(): Promise<HabitsDoc> {
    return (await this.store.get<HabitsDoc>('habits')) ?? { habits: [], completions: [] };
  }

  private view(doc: HabitsDoc, h: HabitRow): Habit {
    const done = new Set(doc.completions.filter((c) => c.habitId === h.id).map((c) => c.day));
    return stats(h, done, isoDay(this.now()));
  }

  private static find(doc: HabitsDoc, id: string): HabitRow {
    const h = doc.habits.find((x) => x.id === id);
    if (!h) throw new Error(`habit ${id} not found`);
    return h;
  }

  async list(): Promise<Habit[]> {
    const doc = await this.read();
    return [...doc.habits]
      .sort((a, b) => a.position - b.position || a.createdAt.localeCompare(b.createdAt))
      .map((h) => this.view(doc, h));
  }

  async create(input: HabitInput): Promise<Habit> {
    const doc = await this.read();
    const row: HabitRow = {
      id: newId(),
      ...input,
      position: doc.habits.reduce((max, h) => Math.max(max, h.position + 1), 0),
      createdAt: this.now().toISOString(),
    };
    doc.habits.push(row);
    await this.store.put('habits', doc);
    return this.view(doc, row);
  }

  async update(id: string, input: HabitInput): Promise<Habit> {
    const doc = await this.read();
    const row = LocalHabitsGateway.find(doc, id);
    Object.assign(row, input);
    await this.store.put('habits', doc);
    return this.view(doc, row);
  }

  async remove(id: string): Promise<void> {
    const doc = await this.read();
    doc.habits = doc.habits.filter((h) => h.id !== id);
    doc.completions = doc.completions.filter((c) => c.habitId !== id);
    await this.store.put('habits', doc);
  }

  async toggle(id: string, date: string): Promise<Habit> {
    const doc = await this.read();
    const row = LocalHabitsGateway.find(doc, id);
    const today = isoDay(this.now());
    if (date !== today && date !== addDays(today, -1)) {
      throw new Error('Only today and yesterday can be checked off.');
    }
    const done = doc.completions.some((c) => c.habitId === id && c.day === date);
    doc.completions = done
      ? doc.completions.filter((c) => !(c.habitId === id && c.day === date))
      : [...doc.completions, { habitId: id, day: date }];
    await this.store.put('habits', doc);
    return this.view(doc, row);
  }
}
