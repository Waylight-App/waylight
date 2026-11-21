import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  CompactAction as CompactActionDTO,
  CompactResponse,
  CompactText as CompactTextDTO,
} from '../../../../shared/api-client/generated/modules';
import {
  type CompactAction,
  compactPath,
  type CompactState,
  type CompactText,
  parseLevel,
} from '../../domain/compact';
import type { CompactGateway } from '../../usecase/compactUseCases';

const toText = (d: CompactTextDTO): CompactText => ({ key: d.key, text: d.text, params: d.params ?? {} });

const toAction = (d: CompactActionDTO): CompactAction => ({
  id: d.id,
  label: toText(d.label),
  method: d.method,
  path: d.path,
  ...(d.body ? { body: d.body } : {}),
  choices: (d.choices ?? []).map((c) => ({
    id: c.id,
    label: toText(c.label),
    ...(c.body ? { body: c.body } : {}),
    selected: c.selected,
  })),
});

/** Maps the DTO of GET /api/<module>/compact. */
export function toCompactState(d: CompactResponse): CompactState {
  return {
    title: toText(d.title),
    ...(d.detail ? { detail: toText(d.detail) } : {}),
    level: parseLevel(d.level),
    ...(d.count !== undefined && d.count !== null ? { count: d.count } : {}),
    updatedAt: d.updatedAt,
    ...(d.action ? { action: toAction(d.action) } : {}),
  };
}

/** CompactGateway backed by the modules' compact endpoints (#597). */
export class CompactApi implements CompactGateway {
  constructor(private readonly http: HttpClient) {}

  async state(source: string): Promise<CompactState> {
    return toCompactState(await this.http.get<CompactResponse>(compactPath(source)));
  }

  async send(method: string, path: string, body?: Record<string, unknown>): Promise<void> {
    switch (method) {
      case 'POST':
        await this.http.post(path, body);
        return;
      case 'PUT':
        await this.http.put(path, body);
        return;
      case 'PATCH':
        await this.http.patch(path, body);
        return;
      case 'DELETE':
        await this.http.delete(path);
        return;
      default:
        throw new Error(`unsupported method ${method}`);
    }
  }
}
