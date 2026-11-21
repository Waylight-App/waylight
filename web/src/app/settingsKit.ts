/**
 * Building blocks of the settings dialog: every feature adds its section from
 * its own src/app/settings/<name>.ts (#193), so the dialog needs no edit.
 */
import type { Component } from 'svelte';
import type { MessageKey } from '../shared/i18n';
import type { Container } from './container';
import type { SettingsArea } from './settingsNav';

/** What the dialog knows besides the container. */
export interface SettingsContext {
  /** Enabled backend modules. */
  modules: ReadonlySet<string>;
}

/** One section of the settings dialog, with its component loaded on demand. */
export interface SettingsSection {
  id: string;
  label: MessageKey;
  /** Shown only while this backend module is enabled. */
  module?: string;
  /** Task area on the settings start page (#600); default "connections". */
  area?: SettingsArea;
  /** Home server or network integration: hidden by the "everyday" profile (#281). */
  homelab?: boolean;
  /** A plugin's own label (ADR 0021); used instead of `label` when set. */
  title?: () => string;
  /** Only admins see it (plugin sections). */
  adminOnly?: boolean;
  /** Contributed by a plugin (ADR 0021): members see it unless adminOnly. */
  plugin?: boolean;
  /** Other words people use for it, in English and German (#601). */
  keywords?: readonly string[];
  /** Labels of fields or headings inside the section the search finds (#601). */
  fields?: readonly MessageKey[];
  load: () => Promise<Component<Record<string, unknown>>>;
  props: (container: Container, context: SettingsContext) => Record<string, unknown>;
}

/**
 * Type-checks the props against the component and erases the type for the
 * registry (the same approach as lazy() in widgetKit.ts).
 */
export function defineSettingsSection<P extends object>(def: {
  id: string;
  label: MessageKey;
  module?: string;
  area?: SettingsArea;
  homelab?: boolean;
  keywords?: readonly string[];
  fields?: readonly MessageKey[];
  load: () => Promise<Component<P>>;
  props: (container: Container, context: SettingsContext) => P;
}): SettingsSection {
  return def as unknown as SettingsSection;
}
