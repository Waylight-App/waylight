/**
 * Named connections of the instance (#610): the address and credential of a
 * service, saved once and chosen by every widget of that service.
 */

/** A credential field a service expects, e.g. apiKey or username. */
export interface CredentialField {
  name: string;
  /** Never shown again once saved; it can only be replaced. */
  secret: boolean;
  optional: boolean;
}

/** A service a connection can be made to (GET /api/integrations/services). */
export interface ConnectionService {
  service: string;
  /** Product name, e.g. "Home Assistant". */
  title: string;
  /** The only address of the service; such connections never ask for one. */
  fixedUrl?: string;
  fields: CredentialField[];
}

/** A widget whose settings name a connection. */
export interface ConnectionUse {
  widgetId: string;
  tabId: string;
  kind: string;
  /** On a tab of the signed-in account; other accounts' tabs stay unnamed. */
  mine: boolean;
}

/** A saved connection without its secrets. */
export interface Connection {
  id: string;
  service: string;
  name: string;
  url: string;
  /** Names of the credential fields that are set. */
  fields: string[];
  /** Credential fields that are not secret, e.g. the user name. */
  values: Record<string, string>;
  usedBy: ConnectionUse[];
}

/** What an admin enters for a new connection. */
export interface ConnectionDraft {
  service: string;
  name: string;
  url: string;
  credential: Record<string, string>;
}

/** The result of the last check of a connection. */
export type ConnectionState =
  | { kind: 'unchecked' }
  | { kind: 'checking' }
  | { kind: 'ok' }
  | { kind: 'failing'; message: string };

/** Connections of one service, as the settings page groups them. */
export interface ServiceGroup {
  service: string;
  title: string;
  connections: Connection[];
}

/** Maximum name length; mirrors the backend. */
export const MAX_NAME_LENGTH = 60;

/** Field names with a label of their own; others are shown as they are. */
const KNOWN_FIELDS = [
  'apiKey',
  'token',
  'username',
  'password',
  'appPassword',
  'tokenId',
  'secret',
  'fingerprint',
] as const;
export type KnownField = (typeof KNOWN_FIELDS)[number];

/** Whether a field has its own label (integrations.field.<name>). */
export function isKnownField(name: string): name is KnownField {
  return (KNOWN_FIELDS as readonly string[]).includes(name);
}

/**
 * Groups connections by service in the order of the service titles, each
 * sorted by name. Connections of a service whose module is switched off
 * keep their service ID as title.
 */
export function groupByService(
  list: readonly Connection[],
  services: readonly ConnectionService[],
  locale?: string,
): ServiceGroup[] {
  const titles = new Map(services.map((s) => [s.service, s.title]));
  const groups = new Map<string, ServiceGroup>();
  for (const c of list) {
    let g = groups.get(c.service);
    if (!g) {
      g = { service: c.service, title: titles.get(c.service) ?? c.service, connections: [] };
      groups.set(c.service, g);
    }
    g.connections.push(c);
  }
  const out = [...groups.values()].sort((a, b) => a.title.localeCompare(b.title, locale));
  for (const g of out) g.connections.sort((a, b) => a.name.localeCompare(b.name, locale));
  return out;
}

/**
 * Validates a draft against its service and returns an error message key,
 * or null when it can be saved.
 */
export function validateDraft(
  d: ConnectionDraft,
  service: ConnectionService | undefined,
):
  | 'integrations.error_service'
  | 'integrations.error_name'
  | 'integrations.error_name_long'
  | 'integrations.error_url'
  | 'integrations.error_field'
  | null {
  if (!service) return 'integrations.error_service';
  const name = d.name.trim();
  if (name === '') return 'integrations.error_name';
  if ([...name].length > MAX_NAME_LENGTH) return 'integrations.error_name_long';
  if (!service.fixedUrl && !/^https?:\/\/\S+$/i.test(d.url.trim())) return 'integrations.error_url';
  if (service.fields.some((f) => !f.optional && !(d.credential[f.name] ?? '').trim())) {
    return 'integrations.error_field';
  }
  return null;
}

/** The credential without empty fields, trimmed; the backend drops empty ones too. */
export function cleanCredential(credential: Record<string, string>): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(credential)) {
    const value = v.trim();
    if (value) out[k] = value;
  }
  return out;
}

/**
 * Describes where a connection is used, e.g. "Immich (Home)": one entry per
 * widget. Widgets on another account's tab are named by kind with `other`.
 */
export function usageList(
  uses: readonly ConnectionUse[],
  widgetName: (kind: string) => string,
  tabName: (tabId: string) => string | undefined,
  other: string,
): string[] {
  return uses.map((u) => {
    const tab = u.mine ? tabName(u.tabId) : other;
    return tab ? `${widgetName(u.kind)} (${tab})` : widgetName(u.kind);
  });
}

/** Counts for the summary line, e.g. "3 working · 1 problem". */
export function stateCounts(states: Iterable<ConnectionState>): { ok: number; failing: number } {
  let ok = 0;
  let failing = 0;
  for (const s of states) {
    if (s.kind === 'ok') ok++;
    if (s.kind === 'failing') failing++;
  }
  return { ok, failing };
}
