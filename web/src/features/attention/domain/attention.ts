/**
 * The attention line of a tab (#593): the most urgent thing that needs a
 * look, lifted out of the widgets, plus how many others there are. Its
 * entries come from two places: the compact states of the tab's status
 * widgets that are amber or red (a service down, a failed backup) and the
 * server's attention list (#599) of connections that stopped working.
 */
export interface AttentionEntry {
  /** Stable key, e.g. "compact:services" or the attention item's ID. */
  id: string;
  /** Problem (red) or attention (amber). */
  level: 'attention' | 'problem';
  /** Who it is about, e.g. the widget's name ("Services"). */
  title: string;
  /** What is wrong, in the viewer's language. */
  text: string;
  /** The widget to show for it, when there is one on this tab. */
  widgetId?: string;
}

const rank = (e: AttentionEntry) => (e.level === 'problem' ? 2 : 1);

/**
 * Orders entries by urgency, problems before attention; entries of the
 * same level keep their order, so widgets come in the order of the tab.
 */
export function byUrgency(entries: readonly AttentionEntry[]): AttentionEntry[] {
  return entries
    .map((e, i) => ({ e, i }))
    .sort((a, b) => rank(b.e) - rank(a.e) || a.i - b.i)
    .map(({ e }) => e);
}

/**
 * The short state of a status widget as the line needs it: its level
 * (none, ok, attention, problem) and its text in the viewer's language.
 */
export interface SourceState {
  level: string;
  text: string;
}

/** A status widget of the tab whose compact state can need attention. */
export interface AttentionSource {
  widgetId: string;
  module: string;
  title: string;
}

/** An entry for a status widget whose state is amber or red; null otherwise. */
export function sourceEntry(source: AttentionSource, state: SourceState): AttentionEntry | null {
  if (state.level !== 'attention' && state.level !== 'problem') return null;
  return {
    id: `compact:${source.module}`,
    level: state.level,
    title: source.title,
    text: state.text,
    widgetId: source.widgetId,
  };
}

/** An item of GET /api/attention as the line needs it. */
export interface AttentionItem {
  id: string;
  /** "connection" (stopped working) or "setup" (never set up). */
  kind: string;
  title: string;
  reason: string;
  widgetId?: string;
  tabId?: string;
}

/**
 * Entries for the server's attention items on a tab: connections that
 * stopped working, on this tab or not tied to any. Missing setup is not a
 * problem of the moment, so the line leaves it to the settings.
 */
export function attentionEntries(items: readonly AttentionItem[], tabId: string): AttentionEntry[] {
  return items
    .filter((i) => i.kind === 'connection' && (!i.tabId || i.tabId === tabId))
    .map((i) => ({
      id: `attention:${i.id}`,
      level: 'problem',
      title: i.title,
      text: i.reason,
      ...(i.widgetId && i.tabId === tabId ? { widgetId: i.widgetId } : {}),
    }));
}
