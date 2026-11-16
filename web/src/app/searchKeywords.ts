import type { KeywordTarget } from '../features/search/domain/keywords';
import type { Container } from './container';
import { widgetDefinitions } from './widgets';

/**
 * What an own keyword (#478) can point to besides web addresses and
 * searches: the service tiles, the bookmarks and the widget kinds. The
 * search feature does not know these features, so the app collects them.
 * A source that fails (module off, no access) just offers nothing.
 */
export async function keywordTargets(c: Container, modules: Set<string>): Promise<KeywordTarget[]> {
  const byLabel = (a: KeywordTarget, b: KeywordTarget) => a.label.localeCompare(b.label);
  const [services, bookmarks] = await Promise.all([
    modules.has('services')
      ? c.services.listAll().then(
          (list) => list.map((s): KeywordTarget => ({ kind: 'service', target: s.id, label: s.name })),
          () => [],
        )
      : [],
    modules.has('bookmarks')
      ? c.bookmarks.list().then(
          (list) =>
            list.bookmarks.map((b): KeywordTarget => ({
              kind: 'bookmark',
              target: b.id,
              label: b.title || b.url,
            })),
          () => [],
        )
      : [],
  ]);
  // A widget keyword opens the widget of that kind on the dashboard, if any.
  const widgets = widgetDefinitions.map((d): KeywordTarget => ({
    kind: 'widget',
    target: d.type.kind,
    label: d.type.title,
  }));
  return [...services.sort(byLabel), ...bookmarks.sort(byLabel), ...widgets.sort(byLabel)];
}
