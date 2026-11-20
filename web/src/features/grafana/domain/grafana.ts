/** Connection settings as shown in the UI (the token is never returned). */
export interface GrafanaSettings {
  configured: boolean;
  url: string;
  hasToken: boolean;
}

/** A dashboard found by the search. */
export interface GrafanaDashboard {
  uid: string;
  title: string;
  folder: string;
  url: string;
}

/** A panel of a dashboard. */
export interface GrafanaPanel {
  id: number;
  title: string;
  type: string;
}

/** A firing or pending alert. */
export interface GrafanaAlert {
  name: string;
  summary: string;
  severity: string;
  state: 'firing' | 'pending';
  since: Date | null;
}

/** Alerts with the link to Grafana's alert list. */
export interface GrafanaAlerts {
  alerts: GrafanaAlert[];
  url: string;
}

/** Settings of the panel widget. */
export interface PanelConfig {
  dashboardUid: string;
  dashboardTitle: string;
  dashboardUrl: string;
  panelId: number;
  panelTitle: string;
  /** Time range back from now, e.g. "6h". */
  range: string;
  /** Minutes between two images. */
  refreshMinutes: number;
}

const RANGE = /^[1-9][0-9]{0,2}[mhdw]$/;

/** Reads the panel widget's settings with defaults (6 h, every 5 minutes). */
export function parsePanelConfig(raw: Record<string, unknown>): PanelConfig {
  const str = (k: string) => (typeof raw[k] === 'string' ? (raw[k] as string) : '');
  const range = str('range');
  const refresh = typeof raw['refreshMinutes'] === 'number' ? Math.round(raw['refreshMinutes']) : 5;
  return {
    dashboardUid: str('dashboardUid'),
    dashboardTitle: str('dashboardTitle'),
    dashboardUrl: str('dashboardUrl'),
    panelId: typeof raw['panelId'] === 'number' && raw['panelId'] > 0 ? raw['panelId'] : 0,
    panelTitle: str('panelTitle'),
    range: RANGE.test(range) ? range : '6h',
    refreshMinutes: Math.min(Math.max(refresh, 1), 60),
  };
}

/** Whether a panel is chosen. */
export function hasPanel(c: PanelConfig): boolean {
  return c.dashboardUid !== '' && c.panelId > 0;
}

/**
 * Address of the rendered panel through Waylight. `stamp` changes the
 * address on each refresh, so the browser fetches a new image.
 */
export function renderUrl(
  c: PanelConfig,
  size: { width: number; height: number },
  dark: boolean,
  stamp: number,
): string {
  const q = new URLSearchParams({
    width: String(Math.round(size.width)),
    height: String(Math.round(size.height)),
    range: c.range,
    theme: dark ? 'dark' : 'light',
    t: String(stamp),
  });
  return `/api/grafana/render/${encodeURIComponent(c.dashboardUid)}/${c.panelId}?${q}`;
}

/** Link to the panel in Grafana, or to the dashboard when the address is unknown. */
export function panelLink(c: PanelConfig): string {
  if (!c.dashboardUrl) return '';
  const sep = c.dashboardUrl.includes('?') ? '&' : '?';
  return `${c.dashboardUrl}${sep}viewPanel=${c.panelId}&from=now-${c.range}&to=now`;
}
