import type { AttentionItem } from '../features/attention/domain/attention';
import type { CatalogSection } from './settingsCatalog';

/** What decides whether the start page points at a hidden advanced section (#622). */
export interface AdvancedHintInput<T extends CatalogSection> {
  /** The attention list of the account (#592). */
  items: readonly AttentionItem[];
  /** Sections the viewer could open with advanced settings shown. */
  sections: readonly T[];
  /** The stored "Show advanced settings" switch. */
  showAdvanced: boolean;
  admin: boolean;
}

/**
 * The advanced section that has a problem while advanced settings are
 * hidden, or undefined. Only entries that stopped working count: a "not set
 * up yet" entry is resolved in its widget, so opening the section would not
 * help. The hint is for admins, who set up the instance's connections.
 */
export function advancedProblem<T extends CatalogSection>(input: AdvancedHintInput<T>): T | undefined {
  if (!input.admin || input.showAdvanced) return undefined;
  const failing = new Set(input.items.filter((i) => i.kind === 'connection').map((i) => i.integration));
  return input.sections.find((s) => s.advanced && !!s.module && failing.has(s.module));
}
