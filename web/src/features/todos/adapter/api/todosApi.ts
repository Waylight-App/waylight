import type { HttpClient } from '../../../../shared/api-client/http';
import type { TodoDTO } from '../../../../shared/api-client/generated/todos';
import type { ClearedDTO, TodoInputDTO } from '../../../../shared/api-client/generated/todos-requests';
import type { DiscoveredListDTO, SourceDTO } from '../../../../shared/api-client/generated/todos';
import type { SourceInputDTO } from '../../../../shared/api-client/generated/todos-requests';
import type { DiscoveredList, Reminder, Repeat, RepeatKind, Todo, TodoSource } from '../../domain/todo';
import type { SourceInput, TodoSourcesGateway, TodosGateway } from '../../usecase/todosUseCases';

const toTodo = (d: TodoDTO): Todo => ({
  id: d.id,
  title: d.title,
  due: d.due,
  done: d.done,
  overdue: d.overdue,
  repeat: {
    kind: (d.repeat?.kind ?? '') as RepeatKind,
    every: d.repeat?.every || 1,
    days: d.repeat?.days ?? [],
  },
  reminder: { at: d.reminder?.at ?? '', daysBefore: d.reminder?.daysBefore ?? 0 },
  version: d.version ?? 0,
});

/** TodosGateway backed by the REST API; source '' is the built-in list (#169). */
export class TodosApi implements TodosGateway {
  constructor(
    private readonly http: HttpClient,
    private readonly source = '',
  ) {}

  /** Adds ?source= (or &source=) for external lists. */
  private url(path: string, query = ''): string {
    const params = new URLSearchParams(query);
    if (this.source) params.set('source', this.source);
    const q = params.toString();
    return q ? `${path}?${q}` : path;
  }

  async list(includeDone: boolean): Promise<Todo[]> {
    return (await this.http.get<TodoDTO[]>(this.url('/api/todos', includeDone ? 'done=true' : ''))).map(
      toTodo,
    );
  }
  async create(title: string, due: string, id?: string): Promise<Todo> {
    // A client-generated id makes a replayed create idempotent (#80).
    const body: TodoInputDTO = { title, due, done: false, ...(id ? { id } : {}) };
    return toTodo(await this.http.post<TodoDTO>(this.url('/api/todos'), body));
  }
  async update(
    id: string,
    title: string,
    due: string,
    done: boolean,
    schedule?: { repeat: Repeat; reminder: Reminder },
    version?: number,
  ): Promise<Todo> {
    // Without a schedule the server keeps the stored recurrence and reminder;
    // with a version it refuses changes made from an older one (409, #80).
    const body: TodoInputDTO = { title, due, done, ...schedule, ...(version ? { version } : {}) };
    return toTodo(await this.http.put<TodoDTO>(this.url(`/api/todos/${encodeURIComponent(id)}`), body));
  }
  delete(id: string): Promise<void> {
    return this.http.delete(this.url(`/api/todos/${encodeURIComponent(id)}`));
  }
  async clearCompleted(): Promise<number> {
    return (await this.http.post<ClearedDTO>(this.url('/api/todos/clear-completed'))).removed;
  }
}

const SOURCES = '/api/todos/settings/sources';

const toSource = (d: SourceDTO): TodoSource => ({
  id: d.id,
  name: d.name,
  kind: 'caldav',
  url: d.url,
  username: d.username,
  capabilities: { ...d.capabilities },
});

/** TodoSourcesGateway backed by the REST API. */
export class TodoSourcesApi implements TodoSourcesGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<TodoSource[]> {
    return (await this.http.get<SourceDTO[]>(SOURCES)).map(toSource);
  }
  async save(id: string, input: SourceInput): Promise<TodoSource> {
    const body: SourceInputDTO = { ...input, kind: 'caldav' };
    return toSource(
      id
        ? await this.http.put<SourceDTO>(`${SOURCES}/${encodeURIComponent(id)}`, body)
        : await this.http.post<SourceDTO>(SOURCES, body),
    );
  }
  delete(id: string): Promise<void> {
    return this.http.delete(`${SOURCES}/${encodeURIComponent(id)}`);
  }
  async discover(input: SourceInput & { id?: string }): Promise<DiscoveredList[]> {
    const body: SourceInputDTO = { ...input, kind: 'caldav' };
    return (await this.http.post<DiscoveredListDTO[]>(`${SOURCES}/discover`, body)).map((l) => ({ ...l }));
  }
}
