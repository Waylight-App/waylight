import type { MessageKey } from '../shared/i18n';
import type { NavArea, SettingsArea } from './settingsNav';

/**
 * The search of the settings (#601): one index of areas, sections and the
 * fields inside sections, built from their definitions. It is built from
 * the sections the viewer may open, so nothing hidden can be found.
 */
export interface SettingsHit {
  kind: 'area' | 'section' | 'field';
  /** Unique within the index, e.g. "field:account:app.change_password". */
  key: string;
  area: SettingsArea;
  /** The section to open; empty for an area. */
  section: string;
  /** What the hit shows, in the UI language. */
  title: string;
  /** Where it lives, e.g. "My account › Account". */
  path: string;
  /** For a field: the label to highlight once the section is open. */
  field?: string;
}

interface Entry extends SettingsHit {
  /** Normalised title and every other word that finds it. */
  terms: string[];
}

export interface SettingsIndex {
  entries: readonly Entry[];
}

export interface IndexSection {
  id: string;
  keywords?: readonly string[];
  fields?: readonly MessageKey[];
}

/** Translations the index needs: the UI language and the other languages. */
export interface IndexTexts<S extends IndexSection> {
  /** A text in the UI language. */
  ui: (key: MessageKey) => string;
  /** The same text in other languages (English and German, #601); may be empty. */
  others: readonly ((key: MessageKey) => string)[];
  /** Display name of a section (plugins have their own labels). */
  sectionName: (s: S) => string;
  sectionLabel: (s: S) => MessageKey | undefined;
}

/** Lower case without accents, so "Zugange" finds "Zugänge" and "e-mail" finds "E-Mail". */
export function normalise(text: string): string {
  return text.normalize('NFD').replace(/\p{M}/gu, '').toLocaleLowerCase().replace(/\s+/g, ' ').trim();
}

/** Builds the index from the visible areas and their sections. */
export function buildIndex<S extends IndexSection>(
  areas: readonly NavArea<S>[],
  texts: IndexTexts<S>,
): SettingsIndex {
  const entries: Entry[] = [];
  const all = (key: MessageKey) => [texts.ui(key), ...texts.others.map((o) => o(key))];
  for (const a of areas) {
    const areaTitle = texts.ui(a.label);
    entries.push({
      kind: 'area',
      key: `area:${a.id}`,
      area: a.id,
      section: '',
      title: areaTitle,
      path: '',
      terms: [...all(a.label), ...a.keywords].map(normalise),
    });
    for (const s of a.sections) {
      const name = texts.sectionName(s);
      const label = texts.sectionLabel(s);
      entries.push({
        kind: 'section',
        key: `section:${s.id}`,
        area: a.id,
        section: s.id,
        title: name,
        path: areaTitle,
        terms: [name, ...(label ? all(label) : []), ...(s.keywords ?? []), s.id.replace(/-/g, ' ')].map(
          normalise,
        ),
      });
      for (const f of s.fields ?? []) {
        const title = texts.ui(f);
        entries.push({
          kind: 'field',
          key: `field:${s.id}:${f}`,
          area: a.id,
          section: s.id,
          title,
          path: `${areaTitle} › ${name}`,
          field: title,
          terms: all(f).map(normalise),
        });
      }
    }
  }
  return { entries };
}

/** Kinds rank areas before sections before fields when they match equally. */
const KIND_ORDER = { area: 0, section: 1, field: 2 } as const;

/**
 * How well the query matches one term: the whole term, its start, the start
 * of a word in it, or anywhere in it (from three characters on, so "ab"
 * does not find every "tab"). 0 means no match.
 */
function termScore(term: string, query: string): number {
  if (term === query) return 4;
  if (term.startsWith(query)) return 3;
  if (term.split(/[\s\-/&›(),.]+/).some((w) => w.startsWith(query))) return 2;
  return query.length >= 3 && term.includes(query) ? 1 : 0;
}

/**
 * The hits for a query, best first. Every word of the query has to match
 * some term of an entry; the entry scores by its weakest word.
 */
export function searchSettings(index: SettingsIndex, query: string, limit = 12): SettingsHit[] {
  const words = normalise(query).split(' ').filter(Boolean);
  if (words.length === 0) return [];
  const scored: { hit: Entry; score: number }[] = [];
  for (const e of index.entries) {
    let score = Infinity;
    for (const w of words) {
      const best = Math.max(...e.terms.map((t) => termScore(t, w)));
      score = Math.min(score, best);
      if (score === 0) break;
    }
    if (score > 0) scored.push({ hit: e, score });
  }
  scored.sort(
    (a, b) =>
      b.score - a.score ||
      KIND_ORDER[a.hit.kind] - KIND_ORDER[b.hit.kind] ||
      a.hit.title.localeCompare(b.hit.title),
  );
  return scored.slice(0, limit).map(({ hit }) => ({
    kind: hit.kind,
    key: hit.key,
    area: hit.area,
    section: hit.section,
    title: hit.title,
    path: hit.path,
    field: hit.field,
  }));
}
