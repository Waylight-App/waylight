/**
 * The plugin API of the web app (ADR 0021). The frontend of ee/ is a separate
 * package that the app loads at runtime; this file is the whole contract
 * between the two. Keep it small and stable: a plugin built against version
 * 1 must keep working while the core changes.
 */

/** Version of this contract; the loader refuses other versions. */
export const PLUGIN_API_VERSION = 1;

/** Access to a licensed feature, as GET /api/licence/features reports it. */
export type FeatureLevel = 'full' | 'read' | 'none';

/** The API client a plugin uses; session and CSRF are handled by the core. */
export interface PluginHttp {
  get<T>(path: string): Promise<T>;
  post<T>(path: string, body?: unknown): Promise<T>;
  put<T>(path: string, body?: unknown): Promise<T>;
  delete<T = void>(path: string): Promise<T>;
  /** PUTs a file as the raw body (images and the like). */
  putBlob<T>(path: string, blob: Blob): Promise<T>;
}

/** What a plugin screen gets from the core besides its own props. */
export interface PluginContext {
  http: PluginHttp;
  /** Active UI language, e.g. "en" or "de". */
  language: string;
  /** The signed-in account. */
  account: { owner: boolean; admin: boolean; member: boolean; guest: boolean };
  /** Access to a licence feature. */
  level(feature: string): FeatureLevel;
  /** Turns a server error into a message for the user, translated. */
  errorMessage(err: unknown): string;
}

/**
 * Something a plugin draws into a host element the core provides. It returns
 * a function that removes it again. Plugins bundle their own Svelte runtime,
 * so they mount into the element instead of joining the core's component tree.
 */
export interface Mountable {
  mount(host: HTMLElement, props: Record<string, unknown>, context: PluginContext): () => void;
}

/** A text in every language the plugin supports; English is the fallback. */
export type Texts = { en: string } & Record<string, string>;

/** A widget kind contributed by a plugin. */
export interface PluginWidget {
  kind: string;
  /** Licence feature that unlocks it. */
  feature: string;
  /** Picker category; organisation features usually belong to "everyday". */
  category: 'everyday' | 'information';
  title: Texts;
  description: Texts;
  keywords?: string[];
  defaultSize: { w: number; h: number };
  /** One of the core's icon names; unknown names fall back to a generic icon. */
  icon: string;
  /** Receives { config }. */
  view: Mountable;
  /** Receives { config, onsave(config) }. */
  settings?: Mountable;
}

/** A section of the settings dialog contributed by a plugin. */
export interface PluginSection {
  id: string;
  /** Licence feature it belongs to; omitted for sections that are always open. */
  feature?: string;
  label: Texts;
  /** Shown instead of the form while the feature is off. */
  description: Texts;
  /** Only admins see it. */
  adminOnly?: boolean;
  view: Mountable;
}

/** What a plugin module exports as `plugin`. */
export interface TrissePlugin {
  apiVersion: number;
  name: string;
  widgets: PluginWidget[];
  settings: PluginSection[];
}

/** Picks the text for a language, falling back to English. */
export const pick = (texts: Texts, language: string): string => texts[language] ?? texts.en;
