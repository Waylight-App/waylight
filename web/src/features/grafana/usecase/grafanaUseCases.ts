import {
  dataUrl,
  type GrafanaAlerts,
  type GrafanaDashboard,
  type GrafanaPanel,
  type GrafanaSettings,
  type PanelConfig,
  type PanelValues,
} from '../domain/grafana';

/** Output port: Grafana backend. */
export interface GrafanaGateway {
  settings(): Promise<GrafanaSettings>;
  saveSettings(url: string, token: string): Promise<GrafanaSettings>;
  alerts(): Promise<GrafanaAlerts>;
  dashboards(query: string): Promise<GrafanaDashboard[]>;
  panels(uid: string): Promise<GrafanaPanel[]>;
  /** The panel's values without the image renderer (#579), by their address. */
  panelValues(url: string): Promise<PanelValues>;
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

  /** The values of the chosen panel, for a stat and sparkline (#579). */
  panelValues(c: PanelConfig): Promise<PanelValues> {
    return this.gateway.panelValues(dataUrl(c));
  }
}
