import { errorMessage, type HttpClient } from '../../shared/api-client/http';
import { locale } from '../../shared/i18n';
import type { IconName } from '../../shared/ui/icons/paths';
import { iconPaths } from '../../shared/ui/icons/paths';
import type { AuthUseCases } from '../../features/auth';
import type { SettingsSection } from '../settingsKit';
import type { WidgetDefinition } from '../widgetKit';
import {
  PLUGIN_API_VERSION,
  pick,
  type FeatureLevel,
  type Mountable,
  type PluginContext,
  type PluginSection,
  type PluginWidget,
  type TrissePlugin,
} from './api';
import LockedFeature from './LockedFeature.svelte';
import PluginHost from './PluginHost.svelte';

/** Where the loaded plugins put their contributions. */
export interface PluginTargets {
  widgets: WidgetDefinition[];
  sections: SettingsSection[];
}

/** How the loader reaches a plugin module; tests replace it. */
export type ImportPlugin = (url: string) => Promise<{ plugin?: TrissePlugin }>;

const browserImport: ImportPlugin = (url) => import(/* @vite-ignore */ url);

/** GETs a JSON endpoint that only exists with ee/; a missing one means "nothing". */
async function optional<T>(http: HttpClient, path: string, fallback: T): Promise<T> {
  try {
    return await http.get<T>(path);
  } catch {
    return fallback;
  }
}

const knownIcon = (name: string): IconName => (name in iconPaths ? (name as IconName) : 'box');

/** Wraps a plugin screen as a lazily "loaded" PluginHost with its props. */
function hosted(m: Mountable, context: PluginContext) {
  return {
    load: () => Promise.resolve(PluginHost as never),
    props: (config: Record<string, unknown>) => ({ mountable: m, context, props: { config } }),
  };
}

function toWidget(w: PluginWidget, context: PluginContext): WidgetDefinition {
  return {
    type: {
      kind: w.kind,
      category: w.category,
      keywords: w.keywords ?? [],
      get title() {
        return pick(w.title, locale());
      },
      get description() {
        return pick(w.description, locale());
      },
      defaultSize: w.defaultSize,
    },
    icon: knownIcon(w.icon),
    view: hosted(w.view, context),
    settings: w.settings ? hosted(w.settings, context) : undefined,
  };
}

/**
 * A plugin section is listed even while its feature is off (ADR 0021): it
 * then shows what the feature does and that it needs a licence key.
 */
function toSection(s: PluginSection, context: PluginContext): SettingsSection {
  const locked = () => !!s.feature && context.level(s.feature) === 'none';
  return {
    id: `plugin-${s.id}`,
    label: 'settings.group.organisation',
    title: () => pick(s.label, locale()),
    group: 'organisation',
    adminOnly: s.adminOnly,
    plugin: true,
    load: () => Promise.resolve((locked() ? LockedFeature : PluginHost) as never),
    props: () =>
      locked()
        ? { title: pick(s.label, locale()), description: pick(s.description, locale()) }
        : { mountable: s.view, context, props: {} },
  };
}

/**
 * Loads the plugins the server offers (GET /api/plugins) and registers their
 * widgets and settings sections. Without ee/ the list is empty and nothing
 * is downloaded. A plugin that fails to load is skipped; the app works on.
 */
export async function loadPlugins(
  http: HttpClient,
  auth: AuthUseCases,
  targets: PluginTargets,
  importPlugin: ImportPlugin = browserImport,
): Promise<string[]> {
  const { plugins } = await optional<{ plugins: string[] }>(http, '/api/plugins', { plugins: [] });
  if (!plugins?.length) return [];
  const { features } = await optional<{ features: Record<string, FeatureLevel> }>(
    http,
    '/api/licence/features',
    {
      features: {},
    },
  );
  const context: PluginContext = {
    http,
    get language() {
      return locale();
    },
    account: { owner: auth.owner, admin: auth.admin, member: auth.member, guest: auth.guest },
    level: (feature) => features[feature] ?? 'none',
    errorMessage,
  };
  const loaded: string[] = [];
  for (const url of plugins) {
    // Only plugins from this server: a relative path, never another origin.
    if (!url.startsWith('/') || url.startsWith('//')) continue;
    try {
      const { plugin } = await importPlugin(url);
      if (!plugin || plugin.apiVersion !== PLUGIN_API_VERSION) {
        console.warn('plugin skipped: unsupported API version', url);
        continue;
      }
      for (const w of plugin.widgets) {
        if (!targets.widgets.some((d) => d.type.kind === w.kind)) targets.widgets.push(toWidget(w, context));
      }
      for (const s of plugin.settings) targets.sections.push(toSection(s, context));
      loaded.push(plugin.name);
    } catch (err) {
      console.error('plugin failed to load', url, err);
    }
  }
  return loaded;
}
