import type { HttpClient } from '../../../../shared/api-client/http';
import type { FoundDTO } from '../../../../shared/api-client/generated/discovery';
import type { FoundService } from '../../../../shared/setup/setup';
import type { DiscoveryGateway } from '../../usecase/discoveryUseCases';

/** DiscoveryGateway backed by GET /api/discovery/{integration}. */
export class DiscoveryApi implements DiscoveryGateway {
  constructor(private readonly http: HttpClient) {}

  async find(integration: string): Promise<FoundService[]> {
    const list = await this.http.get<FoundDTO[]>(`/api/discovery/${encodeURIComponent(integration)}`);
    return list.map((f) => ({ url: f.url, hint: f.hint }));
  }
}
