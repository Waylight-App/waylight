import { withConnection } from '../../../../shared/connection/connection';
import type { HttpClient } from '../../../../shared/api-client/http';
import type { EntityDTO, SettingsDTO } from '../../../../shared/api-client/generated/smarthome';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/smarthome-requests';
import type { Entity, SmartHomeSettings } from '../../domain/smarthome';
import type { SmartHomeGateway } from '../../usecase/smarthomeUseCases';

const ZERO_TIME = '0001-01-01T00:00:00Z';

const toEntity = (d: EntityDTO): Entity => ({
  ...d,
  lastChanged: d.lastChanged && d.lastChanged !== ZERO_TIME ? new Date(d.lastChanged) : null,
});

/** SmartHomeGateway backed by the REST API. */
export class SmarthomeApi implements SmartHomeGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<SmartHomeSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/smarthome/settings')) };
  }
  async saveSettings(url: string, token: string, signIn?: string): Promise<SmartHomeSettings> {
    const body: SettingsInputDTO = { url, token, signIn };
    return { ...(await this.http.put<SettingsDTO>('/api/smarthome/settings', body)) };
  }
  async entities(connection?: string): Promise<Entity[]> {
    return (await this.http.get<EntityDTO[]>(withConnection('/api/smarthome/entities', connection))).map(
      toEntity,
    );
  }
  async states(ids: string[], connection?: string): Promise<Entity[]> {
    const q = new URLSearchParams({ ids: ids.join(',') });
    return (await this.http.get<EntityDTO[]>(withConnection(`/api/smarthome/states?${q}`, connection))).map(
      toEntity,
    );
  }
  act(id: string, connection?: string): Promise<void> {
    return this.http.post(
      withConnection(`/api/smarthome/entities/${encodeURIComponent(id)}/action`, connection),
    );
  }
}
