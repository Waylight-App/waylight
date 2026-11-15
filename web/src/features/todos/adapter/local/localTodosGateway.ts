import { newId } from '../../../../shared/id/uuid';
import type { DocumentStore } from '../../../../shared/local/documentStore';
import type { ShareGroup } from '../../../../shared/sharing/sharing';
import { NO_REMINDER, NO_REPEAT, type Reminder, type Repeat, type Todo } from '../../domain/todo';
import type { TodosGateway } from '../../usecase/todosUseCases';

/** A to-do as stored, in the shape of the server's todos table. */
interface TodoRow {
  id: string;
  title: string;
  /** "YYYY-MM-DD" or '' for none. */
  due: string;
  position: number;
  createdAt: string;
  /** '' while open. */
  completedAt: string;
  repeat: Repeat;
  reminder: Reminder;
  version: number;
}

/** The to-dos document of the local mode. */
export interface TodosDoc {
  todos: TodoRow[];
}

class NotFoundError extends Error {
  override name = 'NotFoundError';
}

/** The local calendar day as "YYYY-MM-DD". */
function dayOf(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

// Dates are handled as UTC midnights, like the server's dateOf, so adding
// days never trips over daylight saving time.
const parse = (s: string) => new Date(`${s}T00:00:00Z`);
const format = (d: Date) => d.toISOString().slice(0, 10);
const addDays = (d: Date, n: number) => new Date(d.getTime() + n * 86_400_000);

/** addMonths keeps the day of month and clamps to the month's last day. */
function addMonths(d: Date, n: number, day: number): Date {
  const first = new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth() + n, 1));
  const last = new Date(Date.UTC(first.getUTCFullYear(), first.getUTCMonth() + 1, 0)).getUTCDate();
  return new Date(Date.UTC(first.getUTCFullYear(), first.getUTCMonth(), Math.min(day, last)));
}

/**
 * Due date of the next occurrence of a recurring to-do, ported from the
 * server's Recurrence.Next (internal/modules/todos/domain/recurrence.go):
 * fixed schedules continue from the due date but never land on or before
 * the day it was done; "after" counts from that day.
 */
export function nextDue(repeat: Repeat, due: string, doneOn: string): string {
  if (repeat.kind === '') return '';
  const done = parse(doneOn);
  if (repeat.kind === 'after') return format(addDays(done, repeat.every));
  const base = due ? parse(due) : done;
  const day = base.getUTCDate();
  const weekStart = (t: Date) => addDays(t, -t.getUTCDay());
  const step = (d: Date): Date => {
    switch (repeat.kind) {
      case 'daily':
        return addDays(d, repeat.every);
      case 'monthly': {
        const months =
          (d.getUTCFullYear() - base.getUTCFullYear()) * 12 +
          (d.getUTCMonth() - base.getUTCMonth()) +
          repeat.every;
        return addMonths(base, months, day);
      }
      default:
        for (let c = addDays(d, 1); ; c = addDays(c, 1)) {
          const weeks = Math.round((weekStart(c).getTime() - weekStart(base).getTime()) / (7 * 86_400_000));
          if (weeks % repeat.every === 0 && repeat.days.includes(c.getUTCDay())) return c;
        }
    }
  };
  let next = base;
  for (let i = 0; i < 5000; i++) {
    next = step(next);
    if (next > done) return format(next);
  }
  return format(addDays(done, 1));
}

/**
 * To-dos gateway of the browser extension's local mode (ADR 0019): the
 * built-in list in one document of the DocumentStore. Completing a recurring
 * to-do creates the next one, as the server does. Reminders are kept with
 * the to-do but not sent: there are no notification channels without a
 * server.
 */
export class LocalTodosGateway implements TodosGateway {
  constructor(
    private readonly store: DocumentStore,
    private readonly now: () => Date = () => new Date(),
  ) {}

  private async read(): Promise<TodosDoc> {
    return (await this.store.get<TodosDoc>('todos')) ?? { todos: [] };
  }

  private async change<T>(fn: (doc: TodosDoc) => T): Promise<T> {
    const doc = await this.read();
    const result = fn(doc);
    await this.store.put('todos', doc);
    return result;
  }

  private toTodo(r: TodoRow): Todo {
    const done = r.completedAt !== '';
    return {
      id: r.id,
      title: r.title,
      due: r.due,
      done,
      overdue: !done && r.due !== '' && r.due < dayOf(this.now()),
      repeat: r.repeat,
      reminder: r.reminder,
      version: r.version,
      visibility: 'private',
      mine: true,
    };
  }

  private static nextPosition(doc: TodosDoc): number {
    return doc.todos.reduce((m, t) => Math.max(m, t.position + 1), 0);
  }

  async list(includeDone: boolean): Promise<Todo[]> {
    const { todos } = await this.read();
    return todos
      .filter((t) => includeDone || t.completedAt === '')
      .sort((a, b) => Number(a.completedAt !== '') - Number(b.completedAt !== '') || a.position - b.position)
      .map((t) => this.toTodo(t));
  }

  create(title: string, due: string, id?: string): Promise<Todo> {
    return this.change((doc) => {
      const row: TodoRow = {
        id: id || newId(),
        title,
        due,
        position: LocalTodosGateway.nextPosition(doc),
        createdAt: this.now().toISOString(),
        completedAt: '',
        repeat: { ...NO_REPEAT, kind: '' },
        reminder: { ...NO_REMINDER },
        version: 1,
      };
      doc.todos.push(row);
      return this.toTodo(row);
    });
  }

  update(
    id: string,
    title: string,
    due: string,
    done: boolean,
    schedule?: { repeat: Repeat; reminder: Reminder },
  ): Promise<Todo> {
    return this.change((doc) => {
      const row = doc.todos.find((t) => t.id === id);
      if (!row) throw new NotFoundError('The to-do does not exist.');
      const completing = done && row.completedAt === '';
      Object.assign(row, { title, due, version: row.version + 1 });
      if (schedule) Object.assign(row, { repeat: schedule.repeat, reminder: schedule.reminder });
      row.completedAt = done ? row.completedAt || this.now().toISOString() : '';
      if (completing && row.repeat.kind !== '') {
        doc.todos.push({
          id: newId(),
          title: row.title,
          due: nextDue(row.repeat, row.due, dayOf(this.now())),
          position: LocalTodosGateway.nextPosition(doc),
          createdAt: this.now().toISOString(),
          completedAt: '',
          repeat: row.repeat,
          reminder: row.reminder,
          version: 1,
        });
      }
      return this.toTodo(row);
    });
  }

  delete(id: string): Promise<void> {
    return this.change((doc) => {
      doc.todos = doc.todos.filter((t) => t.id !== id);
    });
  }

  clearCompleted(): Promise<number> {
    return this.change((doc) => {
      const before = doc.todos.length;
      doc.todos = doc.todos.filter((t) => t.completedAt === '');
      return before - doc.todos.length;
    });
  }

  async share(id: string): Promise<Todo> {
    // One local account: nobody to share with, the to-do stays private.
    const row = (await this.read()).todos.find((t) => t.id === id);
    if (!row) throw new NotFoundError('The to-do does not exist.');
    return this.toTodo(row);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return Promise.resolve([]);
  }
}
