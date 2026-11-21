/**
 * The attention line of a tab (#593): the most urgent thing that needs a
 * look, lifted out of the widgets, plus how many others there are. Its
 * entries come from two places: the short states of the tab's status
 * widgets that are amber or red (a service down, a failed backup) and the
 * account's attention list (#599) of connections that stopped working.
 * The functions take plain values, so the dashboard knows neither the
 * compact views nor the attention feature.
 */

/** One entry of the line. */
export interface AttentionEntry {
  /** Stable key, e.g. "compact:services" or "attention:<id>". */
  id: string;
  /** Problem (red) or attention (amber). */
  level: 'attention' | 'problem';
  /** Who it is about, e.g. the widget's name ("Services"). */
  title: string;
  /** What is wrong, in the viewer's language. */
  text: string;
  /** The widget to show for it, when it is on this tab. */
  widgetId?: string;
}

/** A status widget of the tab and the module that serves its short state. */
export interface StatusSource {
  widgetId: string;
  module: string;
  title: string;
}

/** The short state of a status widget: level (none, ok, attention, problem) and text. */
export interface SourceState {
  level: string;
  text: string;
}

/** An entry for a status widget whose state is amber or red; null otherwise. */
export function sourceEntry(source: StatusSource, state: SourceState): AttentionEntry | null {
  if (state.level !== 'attention' && state.level !== 'problem') return null;
  return {
    id: `compact:${source.module}`,
    level: state.level,
    title: source.title,
    text: state.text,
    widgetId: source.widgetId,
  };
}

/** An item of the attention list as the line needs it. */
export interface ListedProblem {
  id: string;
  /** "connection" (stopped working) or "setup" (never set up). */
  kind: string;
  title: string;
  text: string;
  widgetId?: string;
  tabId?: string;
}

/**
 * Entries for failing connections on this tab or on none. Missing setup is
 * not a problem of the moment, so the line leaves it to the attention page.
 */
export function listedEntries(items: readonly ListedProblem[], tabId: string): AttentionEntry[] {
  return items
    .filter((i) => i.kind === 'connection' && (!i.tabId || i.tabId === tabId))
    .map((i) => ({
      id: `attention:${i.id}`,
      level: 'problem',
      title: i.title,
      text: i.text,
      ...(i.widgetId && i.tabId === tabId ? { widgetId: i.widgetId } : {}),
    }));
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

/** One source per module: the services and status overview widgets share theirs. */
export function uniqueSources(sources: readonly StatusSource[]): StatusSource[] {
  return sources.filter((s, i) => sources.findIndex((o) => o.module === s.module) === i);
}
