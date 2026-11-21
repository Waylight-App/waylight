/**
 * The needs-your-attention list (#592): what stopped working and what is not
 * set up yet, as GET /api/attention serves it (#599).
 */

/** connection = stopped working, setup = not set up yet. */
export type AttentionKind = 'connection' | 'setup';

/** The one thing offered to resolve an entry. */
export type AttentionAction = 'signIn' | 'check' | 'fix' | 'choose' | 'connect';

/** A translatable server message (#102). */
export interface AttentionReason {
  message: string;
  messageKey: string;
  params?: Record<string, string>;
}

/** One entry of the list. */
export interface AttentionItem {
  id: string;
  kind: AttentionKind;
  /** Module the entry belongs to, e.g. "immich". */
  integration: string;
  widgetId?: string;
  tabId?: string;
  reason: AttentionReason;
  since: Date;
  action: AttentionAction;
  /** For "choose": the widget setting that is missing, e.g. "stopId". */
  field?: string;
}

/** The list split the way the page shows it: failures first, then setups. */
export interface AttentionGroups {
  failing: AttentionItem[];
  setup: AttentionItem[];
}

/** Splits the list; the server's order is kept within each group. */
export function groupItems(items: readonly AttentionItem[]): AttentionGroups {
  return {
    failing: items.filter((i) => i.kind === 'connection'),
    setup: items.filter((i) => i.kind === 'setup'),
  };
}

/** Where an action leads. */
export type AttentionTarget =
  | { kind: 'check' }
  | { kind: 'widget'; tabId: string; widgetId: string }
  | { kind: 'settings'; section: string }
  | { kind: 'none' };

/**
 * Decides where an entry's action leads. A missing widget setting opens
 * the widget's own settings; a connection opens the integration's settings
 * section, or the widget when the integration has none; "check" asks the
 * services again in place.
 */
export function actionTarget(
  item: AttentionItem,
  sectionOf: (integration: string) => string | undefined,
): AttentionTarget {
  const widget: AttentionTarget | null =
    item.widgetId && item.tabId ? { kind: 'widget', tabId: item.tabId, widgetId: item.widgetId } : null;
  if (item.action === 'check') return { kind: 'check' };
  if (item.action === 'choose') return widget ?? { kind: 'none' };
  const section = sectionOf(item.integration);
  if (section) return { kind: 'settings', section };
  return widget ?? { kind: 'none' };
}

/** Whether two lists show the same entries, so a poll can skip re-rendering. */
export function sameItems(a: readonly AttentionItem[], b: readonly AttentionItem[]): boolean {
  return (
    a.length === b.length &&
    a.every(
      (x, i) =>
        x.id === b[i]?.id && x.action === b[i]?.action && x.reason.messageKey === b[i]?.reason.messageKey,
    )
  );
}
