import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  AlertsDTO,
  DashboardDTO,
  PanelDTO,
  SettingsDTO,
} from '../../../../shared/api-client/generated/grafana';
import type { SettingsInputDTO } from '../../../../shared/api-client/generated/grafana-requests';
import type { GrafanaAlerts, GrafanaDashboard, GrafanaPanel, GrafanaSettings } from '../../domain/grafana';
import type { GrafanaGateway } from '../../usecase/grafanaUseCases';

/** GrafanaGateway backed by the REST API. */
export class GrafanaApi implements GrafanaGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<GrafanaSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/grafana/settings')) };
  }
  async saveSettings(url: string, token: string): Promise<GrafanaSettings> {
    const body: SettingsInputDTO = { url, token };
    return { ...(await this.http.put<SettingsDTO>('/api/grafana/settings', body)) };
  }
  async alerts(): Promise<GrafanaAlerts> {
    const dto = await this.http.get<AlertsDTO>('/api/grafana/alerts');
    return {
      url: dto.url,
      alerts: dto.alerts.map((a) => ({
        name: a.name,
        summary: a.summary,
        severity: a.severity,
        state: a.state === 'pending' ? 'pending' : 'firing',
        since: a.since ? new Date(a.since) : null,
      })),
    };
  }
  async dashboards(query: string): Promise<GrafanaDashboard[]> {
    const q = new URLSearchParams({ q: query });
    return (await this.http.get<DashboardDTO[]>(`/api/grafana/dashboards?${q}`)).map((d) => ({ ...d }));
  }
  async panels(uid: string): Promise<GrafanaPanel[]> {
    return (await this.http.get<PanelDTO[]>(`/api/grafana/dashboards/${encodeURIComponent(uid)}/panels`)).map(
      (p) => ({ ...p }),
    );
  }
}
