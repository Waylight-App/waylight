import type { GrafanaAlerts, GrafanaDashboard, GrafanaPanel, GrafanaSettings } from '../domain/grafana';

/** Output port: Grafana backend. */
export interface GrafanaGateway {
  settings(): Promise<GrafanaSettings>;
  saveSettings(url: string, token: string): Promise<GrafanaSettings>;
  alerts(): Promise<GrafanaAlerts>;
  dashboards(query: string): Promise<GrafanaDashboard[]>;
  panels(uid: string): Promise<GrafanaPanel[]>;
}

/** Application logic of the Grafana feature (#574). */
export class GrafanaUseCases {
  constructor(private readonly gateway: GrafanaGateway) {}

  settings(): Promise<GrafanaSettings> {
    return this.gateway.settings();
  }

  /** Tests and saves the connection; an empty token keeps the stored one. */
  saveSettings(url: string, token: string): Promise<GrafanaSettings> {
    return this.gateway.saveSettings(url.trim(), token.trim());
  }

  alerts(): Promise<GrafanaAlerts> {
    return this.gateway.alerts();
  }

  dashboards(query: string): Promise<GrafanaDashboard[]> {
    return this.gateway.dashboards(query.trim());
  }

  panels(uid: string): Promise<GrafanaPanel[]> {
    return this.gateway.panels(uid);
  }
}
