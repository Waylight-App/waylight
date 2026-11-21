/**
 * In-app help (#678): the short help of each widget kind and plain-language
 * questions answered from it and from the documentation index. Pure
 * functions; the texts come from the server so the app shows the same.
 */

/** The help of one widget kind, already in the current language. */
export interface WidgetHelp {
  kind: string;
  shows: string;
  setUp: string;
  docsUrl: string;
}

/** Something a question can lead to: a widget's help or a documentation page. */
export interface HelpTopic {
  id: string;
  title: string;
  /** One line under the title. */
  detail: string;
  /** Extra search terms, such as the widget picker's keywords. */
  keywords: string[];
  /** The widget kind whose help opens; absent for documentation pages. */
  kind?: string;
  /** Where the topic leads outside the app. */
  url: string;
  /** Topics of one group (the sections of a page) are listed once, the best one. */
  group?: string;
}

/** A documentation page as the index lists it. */
export interface DocsPage {
  path: string;
  title: string;
  summary: string;
  sections: { title: string; anchor: string }[];
}

/** Words that carry no meaning in a question, in English and German. */
const STOP_WORDS = new Set(
  (
    'a an and are can could do does for from get how i in is it me my of on or show see set the this to up use ' +
    'what when where which who why will with you your ' +
    'auf aus bei das dem den der die ein eine einen einem einer es für gibt hinzu ich im in ist kann man mein ' +
    'meine meinen mir mit nach oder sehe sich sie und was welche wie wo warum wieso zu zeigt'
  ).split(' '),
);

/** Words that start a question. */
const QUESTION_START =
  /^(how|what|where|why|which|can|could|do|does|is|wie|was|wo|warum|wieso|welche[rsn]?|kann|gibt)\b/;

/** Lower case without accents, so "Ähnlich" matches "ahnlich". */
export function normalise(s: string): string {
  return s
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase();
}

/**
 * Whether the text reads as a question to the help: it ends with a question
 * mark, starts with one (the help prefix) or starts with a question word.
 */
export function isQuestion(query: string): boolean {
  const q = normalise(query.trim());
  if (q.length < 4) return false;
  return q.endsWith('?') || q.startsWith('?') || QUESTION_START.test(q);
}

/**
 * The meaningful words of a question, cut to a stem so plural and inflected
 * forms match ("calendars" → "calenda", "Kalender" → "kalend").
 */
export function questionWords(query: string): string[] {
  return normalise(query)
    .split(/[^\p{L}\p{N}]+/u)
    .filter((w) => w.length > 1 && !STOP_WORDS.has(w))
    .map((w) => (w.length > 5 ? w.slice(0, Math.max(5, w.length - 2)) : w));
}

/**
 * The topics that answer a question, best first. A topic has to match at
 * least one word in its title or keywords; words in the detail only rank
 * it higher, so a word like "add" alone finds nothing.
 */
export function matchTopics(query: string, topics: readonly HelpTopic[], limit = 4): HelpTopic[] {
  const words = questionWords(query.replace(/^\?/, ''));
  if (words.length === 0) return [];
  const scored: { topic: HelpTopic; score: number; i: number }[] = [];
  topics.forEach((topic, i) => {
    const title = normalise(topic.title);
    const keywords = topic.keywords.map(normalise);
    const detail = normalise(topic.detail);
    let strong = 0;
    let score = 0;
    for (const w of words) {
      if (title.includes(w)) {
        strong++;
        score += 3;
      } else if (keywords.some((k) => k.includes(w))) {
        strong++;
        score += 2;
      } else if (detail.includes(w)) score += 1;
    }
    if (strong > 0) scored.push({ topic, score, i });
  });
  const seen = new Set<string>();
  return scored
    .sort((a, b) => b.score - a.score || a.i - b.i)
    .map((s) => s.topic)
    .filter((t) => {
      if (t.group === undefined) return true;
      if (seen.has(t.group)) return false;
      seen.add(t.group);
      return true;
    })
    .slice(0, limit);
}

/** Topics of the widgets: their help in the app comes before the documentation. */
export function widgetTopics(
  help: readonly WidgetHelp[],
  describe: (kind: string) => { title: string; keywords: string[] } | undefined,
): HelpTopic[] {
  return help.flatMap((h) => {
    const d = describe(h.kind);
    if (!d) return [];
    return [
      {
        id: `widget:${h.kind}`,
        title: d.title,
        detail: h.shows,
        keywords: [...d.keywords, h.setUp],
        kind: h.kind,
        url: h.docsUrl,
      },
    ];
  });
}

/** Topics of the documentation: each page and each of its sections. */
export function docsTopics(baseUrl: string, pages: readonly DocsPage[]): HelpTopic[] {
  return pages.flatMap((p) => [
    {
      id: `docs:${p.path}`,
      title: p.title,
      detail: p.summary,
      keywords: [],
      url: baseUrl + p.path,
      group: p.path,
    },
    ...p.sections.map((s) => ({
      id: `docs:${p.path}#${s.anchor}`,
      title: s.title,
      detail: p.title,
      keywords: [p.title],
      url: `${baseUrl}${p.path}#${s.anchor}`,
      group: p.path,
    })),
  ]);
}
