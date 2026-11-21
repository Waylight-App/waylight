import type { HttpClient } from '../../../../shared/api-client/http';
import { serverMessage } from '../../../../shared/i18n';
import type { AttentionItem } from '../../domain/attention';
import type { AttentionGateway } from '../../usecase/attentionUseCases';

/** The answer of GET /api/attention (see api/modules/attention.yaml). */
interface ListDTO {
  items: {
    id: string;
    kind: string;
    title: string;
    reason: { message: string; messageKey: string; params?: Record<string, string> };
    widgetId?: string;
    tabId?: string;
  }[];
}

/** AttentionGateway backed by GET /api/attention. */
export class AttentionApi implements AttentionGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<AttentionItem[]> {
    const dto = await this.http.get<ListDTO>('/api/attention');
    return (dto.items ?? []).map((i) => ({
      id: i.id,
      kind: i.kind,
      title: i.title,
      reason: serverMessage(i.reason.messageKey, i.reason.params) ?? i.reason.message,
      ...(i.widgetId ? { widgetId: i.widgetId } : {}),
      ...(i.tabId ? { tabId: i.tabId } : {}),
    }));
  }
}
