import type { HttpClient } from '../../../../shared/api-client/http';
import type { ConnectionDTO, ListDTO } from '../../../../shared/api-client/generated/attention';
import type {
  AttentionAction,
  AttentionItem,
  AttentionKind,
  ServiceConnection,
} from '../../domain/attention';
import type { AttentionGateway } from '../../usecase/attentionUseCases';

/** AttentionGateway backed by GET /api/attention (#599). */
export class AttentionApi implements AttentionGateway {
  constructor(private readonly http: HttpClient) {}

  async list(fresh: boolean): Promise<AttentionItem[]> {
    const d = await this.http.get<ListDTO>(fresh ? '/api/attention?refresh=true' : '/api/attention');
    return d.items.map((i) => ({
      id: i.id,
      kind: i.kind as AttentionKind,
      integration: i.integration,
      widgetId: i.widgetId,
      tabId: i.tabId,
      reason: { message: i.reason.message, messageKey: i.reason.messageKey, params: i.reason.params },
      since: new Date(i.since),
      action: i.action as AttentionAction,
      field: i.field,
    }));
  }

  async connections(fresh: boolean): Promise<ServiceConnection[]> {
    const d = await this.http.get<ConnectionDTO[]>(
      fresh ? '/api/attention/connections?refresh=true' : '/api/attention/connections',
    );
    return d.map((c) => ({
      integration: c.integration,
      instance: c.instance,
      state: c.state as ServiceConnection['state'],
      reason: c.reason ? { ...c.reason } : undefined,
      action: c.action ? (c.action as AttentionAction) : undefined,
      widgets: c.widgets.map((w) => ({ ...w })),
    }));
  }
}
