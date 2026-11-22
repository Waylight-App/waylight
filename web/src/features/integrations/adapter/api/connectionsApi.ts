import type { HttpClient } from '../../../../shared/api-client/http';
import type { ConnectionChoice } from '../../../../shared/connection/connection';
import type {
  ChoiceDTO,
  ConnectionDTO,
  ServiceDTO,
} from '../../../../shared/api-client/generated/integrations';
import type {
  ConnectionChangeDTO,
  ConnectionInputDTO,
  CredentialDTO,
} from '../../../../shared/api-client/generated/integrations-requests';
import type { Connection, ConnectionDraft, ConnectionService } from '../../domain/integrations';
import type { ConnectionChange, ConnectionsGateway } from '../../usecase/connectionsUseCases';

const toService = (d: ServiceDTO): ConnectionService => ({
  service: d.service,
  title: d.title,
  fixedUrl: d.fixedUrl,
  fields: d.fields.map((f) => ({ ...f })),
});

const toConnection = (d: ConnectionDTO): Connection => ({
  id: d.id,
  service: d.service,
  name: d.name,
  url: d.url,
  fields: [...d.fields],
  values: { ...d.values },
  usedBy: d.usedBy.map((u) => ({ ...u })),
});

const path = (id: string) => `/api/integrations/connections/${encodeURIComponent(id)}`;

/** ConnectionsGateway backed by the REST API (#610). */
export class ConnectionsApi implements ConnectionsGateway {
  constructor(private readonly http: HttpClient) {}

  async services(): Promise<ConnectionService[]> {
    return (await this.http.get<ServiceDTO[]>('/api/integrations/services')).map(toService);
  }

  async list(): Promise<Connection[]> {
    return (await this.http.get<ConnectionDTO[]>('/api/integrations/connections')).map(toConnection);
  }

  async choices(service: string): Promise<ConnectionChoice[]> {
    const q = new URLSearchParams({ service });
    return (await this.http.get<ChoiceDTO[]>(`/api/integrations/choices?${q}`)).map((d) => ({ ...d }));
  }

  async add(d: ConnectionDraft): Promise<Connection> {
    const body: ConnectionInputDTO = { ...d };
    return toConnection(await this.http.post<ConnectionDTO>('/api/integrations/connections', body));
  }

  async update(id: string, change: ConnectionChange): Promise<Connection> {
    const body: ConnectionChangeDTO = { ...change };
    return toConnection(await this.http.put<ConnectionDTO>(path(id), body));
  }

  async replaceCredential(id: string, credential: Record<string, string>): Promise<Connection> {
    const body: CredentialDTO = { credential };
    return toConnection(await this.http.put<ConnectionDTO>(`${path(id)}/credential`, body));
  }

  check(id: string): Promise<void> {
    return this.http.post(`${path(id)}/check`);
  }

  remove(id: string): Promise<void> {
    // The settings page asks for confirmation first, also for a connection in use.
    return this.http.delete(`${path(id)}?confirm=true`);
  }
}
