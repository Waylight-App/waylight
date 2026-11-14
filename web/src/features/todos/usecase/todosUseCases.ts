import {
  type DiscoveredList,
  type Reminder,
  type SourceKind,
  type Repeat,
  type Todo,
  type TodoSource,
  validateSchedule,
} from '../domain/todo';
import { t } from '../../../shared/i18n';

/** Output port: to-do backend. */
export interface TodosGateway {
  list(includeDone: boolean): Promise<Todo[]>;
  /** id: client-generated, for idempotent replays (#80). */
  create(title: string, due: string, id?: string): Promise<Todo>;
  update(
    id: string,
    title: string,
    due: string,
    done: boolean,
    schedule?: { repeat: Repeat; reminder: Reminder },
    /** The version the change was made from; 0/undefined = no check (#80). */
    version?: number,
  ): Promise<Todo>;
  delete(id: string): Promise<void>;
  clearCompleted(): Promise<number>;
}

/** Application logic of the to-do feature. */
export class TodosUseCases {
  constructor(private readonly gateway: TodosGateway) {}

  list(includeDone: boolean): Promise<Todo[]> {
    return this.gateway.list(includeDone);
  }

  /**
   * Adds a to-do. A trailing "@YYYY-MM-DD" in the text sets the due date,
   * e.g. "Pay rent @2026-12-01".
   */
  add(text: string): Promise<Todo> {
    const match = /\s@(\d{4}-\d{2}-\d{2})\s*$/.exec(text);
    const title = match ? text.slice(0, match.index) : text;
    return this.gateway.create(title.trim(), match?.[1] ?? '');
  }

  toggle(todo: Todo): Promise<Todo> {
    return this.gateway.update(todo.id, todo.title, todo.due, !todo.done);
  }

  edit(todo: Todo, title: string, due: string): Promise<Todo> {
    return this.gateway.update(todo.id, title.trim(), due, todo.done);
  }

  /** Changes text, due date, recurrence and reminder (#68). */
  async editFull(
    todo: Todo,
    change: { title: string; due: string; repeat: Repeat; reminder: Reminder },
  ): Promise<Todo> {
    const repeat: Repeat =
      change.repeat.kind === ''
        ? { kind: '', every: 0, days: [] }
        : {
            ...change.repeat,
            days: change.repeat.kind === 'weekly' ? [...new Set(change.repeat.days)].sort() : [],
          };
    const problem = validateSchedule(change.due, change.repeat, change.reminder);
    if (problem) throw new Error(problem);
    return this.gateway.update(todo.id, change.title.trim(), change.due, todo.done, {
      repeat,
      reminder: change.reminder,
    });
  }

  delete(id: string): Promise<void> {
    return this.gateway.delete(id);
  }

  clearCompleted(): Promise<number> {
    return this.gateway.clearCompleted();
  }
}

/**
 * A list to connect or change; an empty password keeps the stored one.
 * Todoist needs only the API token (password) and a project (listId).
 */
export interface SourceInput {
  kind: SourceKind;
  name: string;
  url: string;
  username: string;
  password: string;
  listId?: string;
}

/** Output port: connected external lists (#169). */
export interface TodoSourcesGateway {
  list(): Promise<TodoSource[]>;
  /** Creates (id '') or changes a list; the server tests the connection. */
  save(id: string, input: SourceInput): Promise<TodoSource>;
  delete(id: string): Promise<void>;
  discover(input: SourceInput & { id?: string }): Promise<DiscoveredList[]>;
}

/** Connecting external to-do lists (CalDAV, #169). */
export class TodoSourcesUseCases {
  private cached: Promise<TodoSource[]> | null = null;

  constructor(private readonly gateway: TodoSourcesGateway) {}

  /** The connected lists; asked once, guests (no access) get none. */
  list(fresh = false): Promise<TodoSource[]> {
    if (fresh || !this.cached) this.cached = this.gateway.list().catch(() => []);
    return this.cached;
  }

  private check(input: SourceInput, needsPassword: boolean): void {
    if (input.kind === 'todoist') {
      if (needsPassword && !input.password.trim()) throw new Error(t('todos.sources.token_needed'));
      return;
    }
    if (!/^https?:\/\/\S+/i.test(input.url.trim())) throw new Error(t('todos.sources.url_needed'));
    if (!input.username.trim() || (needsPassword && !input.password))
      throw new Error(t('todos.sources.login_needed'));
  }

  /** Finds the task lists of an account (from the DAV address or a list). */
  async discover(input: SourceInput, id = ''): Promise<DiscoveredList[]> {
    this.check(input, !id);
    return this.gateway.discover({ ...input, url: input.url.trim(), ...(id ? { id } : {}) });
  }

  async save(id: string, input: SourceInput): Promise<TodoSource> {
    if (!input.name.trim()) throw new Error(t('todos.sources.name_needed'));
    this.check(input, !id);
    if (input.kind === 'todoist' && !input.listId) throw new Error(t('todos.sources.project_needed'));
    const s = await this.gateway.save(id, { ...input, name: input.name.trim(), url: input.url.trim() });
    this.cached = null;
    return s;
  }

  async delete(id: string): Promise<void> {
    await this.gateway.delete(id);
    this.cached = null;
  }
}
