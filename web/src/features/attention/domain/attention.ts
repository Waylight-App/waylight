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

/** How one integration's connection is doing (#591). */
export interface ServiceConnection {
  integration: string;
  /** A connection of the instance, which admins set up. */
  instance: boolean;
  state: 'ok' | 'setup' | 'failing';
  reason?: AttentionReason;
  action?: AttentionAction;
  widgets: { id: string; tabId: string; kind: string }[];
}

/** Counts for a summary line such as "9 connected · 1 problem". */
export function connectionCounts(list: readonly ServiceConnection[]): {
  connected: number;
  failing: number;
  setup: number;
} {
  return {
    connected: list.filter((c) => c.state === 'ok').length,
    failing: list.filter((c) => c.state === 'failing').length,
    setup: list.filter((c) => c.state === 'setup').length,
  };
}

/**
 * Orders connections for the page: problems first, then what is connected,
 * then what can still be connected, each by name.
 */
export function sortConnections(
  list: readonly ServiceConnection[],
  name: (integration: string) => string,
  locale?: string,
): ServiceConnection[] {
  const rank = { failing: 0, ok: 1, setup: 2 } as const;
  return [...list].sort(
    (a, b) => rank[a.state] - rank[b.state] || name(a.integration).localeCompare(name(b.integration), locale),
  );
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
