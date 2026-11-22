import { connectionIdOf, withConnection } from '../../../shared/connection/connection';

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
  /** Named connection to read through (#709); empty reads the module settings. */
  connectionId: string;
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
    connectionId: connectionIdOf(raw),
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
  return withConnection(
    `/api/grafana/render/${encodeURIComponent(c.dashboardUid)}/${c.panelId}?${q}`,
    c.connectionId,
  );
}

/** Link to the panel in Grafana, or to the dashboard when the address is unknown. */
export function panelLink(c: PanelConfig): string {
  if (!c.dashboardUrl) return '';
  const sep = c.dashboardUrl.includes('?') ? '&' : '?';
  return `${c.dashboardUrl}${sep}viewPanel=${c.panelId}&from=now-${c.range}&to=now`;
}

/** One line of a panel's values, oldest point first (#579). */
export interface PanelSeries {
  name: string;
  /** `t` is null for a value without time (a reduced value). */
  points: { t: Date | null; v: number }[];
}

/** The values of a panel, read without Grafana's image renderer (#579). */
export interface PanelValues {
  title: string;
  /** Grafana's unit id, e.g. "percent"; "" without one. */
  unit: string;
  decimals?: number;
  series: PanelSeries[];
}

/** The newest value of a series, or null for an empty one. */
export function latest(s: PanelSeries | undefined): number | null {
  const last = s?.points.at(-1);
  return last ? last.v : null;
}

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

/** Grafana units that are a plain suffix. */
const SUFFIXES: Record<string, string> = {
  percent: ' %',
  celsius: ' °C',
  fahrenheit: ' °F',
  humidity: ' %',
  watt: ' W',
  kwatt: ' kW',
  kwatth: ' kWh',
  watth: ' Wh',
  volt: ' V',
  amp: ' A',
  hertz: ' Hz',
  s: ' s',
  ms: ' ms',
  m: ' min',
  h: ' h',
  d: ' d',
  bps: ' b/s',
  Bps: ' B/s',
  reqps: ' req/s',
  ops: ' ops/s',
  lengthm: ' m',
  lengthkm: ' km',
  velocitykmh: ' km/h',
  pressurehpa: ' hPa',
  currencyEUR: ' €',
  currencyUSD: ' $',
};

/**
 * Formats a panel value with the panel's unit the way Grafana roughly would:
 * the common units, byte sizes scaled, and custom "suffix:" or "prefix:"
 * units; an unknown unit shows the bare number rather than a wrong unit.
 */
export function formatPanelValue(
  v: number,
  unit: string,
  decimals: number | undefined,
  locale: string,
): string {
  let value = v;
  let prefix = '';
  let suffix = '';
  if (unit === 'percentunit') {
    value = v * 100;
    suffix = ' %';
  } else if (unit === 'bytes' || unit === 'decbytes') {
    const base = unit === 'bytes' ? 1024 : 1000;
    let i = 0;
    while (Math.abs(value) >= base && i < BYTE_UNITS.length - 1) {
      value /= base;
      i++;
    }
    suffix = ` ${BYTE_UNITS[i]}`;
  } else if (unit.startsWith('suffix:')) {
    suffix = unit.slice('suffix:'.length);
  } else if (unit.startsWith('prefix:')) {
    prefix = unit.slice('prefix:'.length);
  } else {
    suffix = SUFFIXES[unit] ?? '';
  }
  // Without the panel's decimals: whole numbers stay whole, small ones keep
  // enough digits to be readable.
  const digits = decimals ?? (Number.isInteger(value) ? 0 : Math.abs(value) < 10 ? 2 : 1);
  const text = new Intl.NumberFormat(locale, {
    maximumFractionDigits: digits,
    minimumFractionDigits: 0,
  }).format(value);
  return `${prefix}${text}${suffix}`;
}

/** Address of the panel's values through Waylight (#579). */
export function dataUrl(c: PanelConfig): string {
  const path = `/api/grafana/data/${encodeURIComponent(c.dashboardUid)}/${c.panelId}?${new URLSearchParams({ range: c.range })}`;
  return withConnection(path, c.connectionId);
}
