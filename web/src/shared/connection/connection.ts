/**
 * The named connection a widget reads through (#610): the widget stores its
 * ID as `connectionId` (`connectionIds` for several) and its requests carry
 * it as `?connection=<id>`. Without a choice the module settings are read.
 */

/** A saved connection as a widget's picker lists it: no address, no access details. */
export interface ConnectionChoice {
  id: string;
  /** Service key, e.g. "immich". */
  service: string;
  name: string;
}

/** Loads the connections of one service. */
export type ChoiceLoader = (service: string) => Promise<ConnectionChoice[]>;

/** The saved connections a widget offers in its setup, and how a choice is stored. */
export interface SavedConnections {
  /** Service keys of the widget, e.g. ['immich']. */
  services: string[];
  load: ChoiceLoader;
  /** Stores the chosen connection in the widget. */
  choose: (id: string) => Promise<void>;
}

/** Appends the chosen connection to a request path; none leaves it as it is. */
export function withConnection(path: string, id?: string): string {
  return withConnections(path, id ? [id] : []);
}

/** Appends several chosen connections, repeated as the API expects. */
export function withConnections(path: string, ids: readonly string[] = []): string {
  if (ids.length === 0) return path;
  const query = ids.map((id) => `connection=${encodeURIComponent(id)}`).join('&');
  return `${path}${path.includes('?') ? '&' : '?'}${query}`;
}

/** The connection named in a widget's configuration, or '' for the module settings. */
export function connectionIdOf(raw: Record<string, unknown>): string {
  const id = raw['connectionId'];
  return typeof id === 'string' ? id : '';
}

/** Query parameters of a compact source (#638) for the connection a widget chose. */
export function connectionParams(raw: Record<string, unknown>): Record<string, string> {
  const connection = connectionIdOf(raw);
  return connection ? { connection } : {};
}

/** The connections named in a widget's configuration. */
export function connectionIdsOf(raw: Record<string, unknown>): string[] {
  const ids = raw['connectionIds'];
  return Array.isArray(ids) ? ids.filter((id): id is string => typeof id === 'string' && id !== '') : [];
}

/**
 * The widget configuration with the choice set; an empty choice removes the
 * key, so the widget goes back to the module settings.
 */
export function withConnectionChoice(config: Record<string, unknown>, id: string): Record<string, unknown> {
  const rest = { ...config };
  delete rest['connectionId'];
  return id ? { ...rest, connectionId: id } : rest;
}

/**
 * What a widget's setup offers: the saved connections of its services, and
 * storing the choice in its configuration. Undefined when the widget cannot
 * save (no loader or no `onsave`), so the setup then asks for an address only.
 */
export function savedConnections(
  services: string[],
  load: ChoiceLoader | undefined,
  config: Record<string, unknown>,
  save: ((config: Record<string, unknown>) => Promise<void>) | undefined,
): SavedConnections | undefined {
  if (!load || !save) return undefined;
  return { services, load, choose: (id) => save(withConnectionChoice(config, id)) };
}
