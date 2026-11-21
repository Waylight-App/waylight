/**
 * The attention list of the signed-in account (#592), shared by the top bar
 * dot, the settings page and the settings start page (#600). The shell
 * attaches it to the server and says how actions are carried out; this
 * file imports neither the settings nor the widgets, so the settings
 * section can use it without an import cycle.
 */
// Imported file by file, not through the feature's index, so the page
// stays out of the first chunk (ADR 0011).
import { AttentionState } from '../features/attention/ui/attentionState.svelte';
import {
  actionTarget,
  type AttentionItem,
  type AttentionTarget,
} from '../features/attention/domain/attention';

export const attention = new AttentionState();

/** What the shell knows: where things are and how to open them. */
export interface AttentionHandlers {
  openWidget: (tabId: string, widgetId: string) => void;
  openSettings: (section: string) => void;
  /** The settings section of an integration, e.g. "immich" → "immich". */
  sectionOf: (integration: string) => string | undefined;
  /** Display name of an integration. */
  nameOf: (integration: string) => string;
  /** Name of a tab of the account. */
  tabName: (tabId: string) => string | undefined;
}

let handlers: AttentionHandlers | null = null;

/** Called by the shell once it can open widgets and settings sections. */
export function setAttentionHandlers(h: AttentionHandlers | null): void {
  handlers = h;
}

/** Display name of an integration; the module name until the shell is ready. */
export function integrationName(integration: string): string {
  return handlers?.nameOf(integration) ?? integration;
}

/** Name of a tab, to say where a widget is. */
export function tabName(tabId: string): string | undefined {
  return handlers?.tabName(tabId);
}

/** Carries out an entry's action and returns where it led. */
export function runAttentionAction(item: AttentionItem): AttentionTarget {
  const target = actionTarget(item, (i) => handlers?.sectionOf(i));
  switch (target.kind) {
    case 'check':
      void attention.check();
      break;
    case 'widget':
      handlers?.openWidget(target.tabId, target.widgetId);
      break;
    case 'settings':
      handlers?.openSettings(target.section);
      break;
  }
  return target;
}
