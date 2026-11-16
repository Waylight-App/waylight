import { type CommandSpec, permitted, type Viewer } from '../features/search/domain/commands';
import type { HelpEntry } from '../features/search/domain/help';
import type { IntentSpec } from '../features/search/domain/smart/intent';
import { INTENTS } from '../features/search/domain/smart/intents';
import { locale, type MessageKey, t } from '../shared/i18n';
import { contributions } from './searchSources';

/**
 * Builds the search help (#479) from the data that drives the search, in
 * the current language: the intents of the enabled modules, the app
 * commands the account may run, the capture prefixes and the answers that
 * need no module (calculator, time). Nothing here repeats what an intent
 * or command already says about itself.
 */

type Lang = 'de' | 'en';
type Examples = Record<Lang, string[]>;

/** The module that has to be enabled for an intent; core intents use their source. */
const MODULE_OF = new Map<string, string>([
  ...INTENTS.map((i) => [i.id, i.source] as const),
  ...contributions.flatMap((c) => c.intents.map((i) => [i.id, c.module] as const)),
]);

const ALL_INTENTS: IntentSpec[] = [...INTENTS, ...contributions.flatMap((c) => c.intents)];

/** Intents that change something; guests and kiosk displays cannot run them. */
const ACTING = new Set(['docker.restart']);

/** Answers without a module and the web search; examples per language. */
const FIXED: {
  id: string;
  group: HelpEntry['group'];
  text: MessageKey;
  examples: Examples;
  module?: string;
}[] = [
  {
    id: 'time',
    group: 'time',
    text: 'search.help.time',
    examples: {
      de: ['kw', 'zeit in tokio', 'tage bis weihnachten'],
      en: ['week', 'time in tokyo', 'days until christmas'],
    },
  },
  {
    id: 'calc',
    group: 'tools',
    text: 'search.help.calc',
    examples: { de: ['19 % von 240', '5 km in mi'], en: ['19 % of 240', '5 km in mi'] },
  },
  {
    id: 'web',
    group: 'web',
    text: 'search.help.web',
    module: 'search',
    examples: { de: ['raclette rezept'], en: ['raclette recipe'] },
  },
  {
    id: 'bang',
    group: 'web',
    text: 'search.help.bang',
    module: 'search',
    examples: { de: ['!w berlin'], en: ['!w berlin'] },
  },
];

/** Quick capture: the module, its help text and an example after the prefix. */
const CAPTURE: { kind: 'todo' | 'note' | 'bookmark'; module: string; text: MessageKey; rest: Examples }[] = [
  {
    kind: 'todo',
    module: 'todos',
    text: 'search.help.capture_todo',
    rest: { de: ['Bank anrufen morgen'], en: ['call the bank tomorrow'] },
  },
  {
    kind: 'note',
    module: 'notes',
    text: 'search.help.capture_note',
    rest: { de: ['Idee für den Garten'], en: ['idea for the garden'] },
  },
  {
    kind: 'bookmark',
    module: 'bookmarks',
    text: 'search.help.capture_bookmark',
    rest: { de: ['https://example.org'], en: ['https://example.org'] },
  },
];

export interface HelpInput {
  modules: Set<string>;
  viewer: Viewer;
  /** Capture prefixes in use, e.g. { todo: ["t"] }. */
  prefixes: Partial<Record<'todo' | 'note' | 'bookmark', string[]>>;
  /** The app commands of the shell (#476); none on pages without them. */
  commands?: { specs: () => CommandSpec[]; title: (spec: CommandSpec) => string };
}

/** Every help entry for the account and the enabled modules, in the current language. */
export function buildHelp(input: HelpInput): HelpEntry[] {
  const lang: Lang = locale() === 'de' ? 'de' : 'en';
  const { modules, viewer } = input;
  const out: HelpEntry[] = [];
  const reader = viewer.guest || viewer.kiosk;

  for (const intent of ALL_INTENTS) {
    const module = MODULE_OF.get(intent.id);
    if (!intent.help || !module || !modules.has(module)) continue;
    if (reader && ACTING.has(intent.id)) continue;
    out.push({
      id: intent.id,
      group: intent.help.group,
      text: intent.help[lang],
      examples: intent.help.examples[lang],
    });
  }

  for (const f of FIXED) {
    if (f.module && !modules.has(f.module)) continue;
    out.push({ id: f.id, group: f.group, text: t(f.text), examples: f.examples[lang] });
  }

  if (!reader) {
    for (const c of CAPTURE) {
      const prefix = input.prefixes[c.kind]?.[0];
      if (!prefix || !modules.has(c.module)) continue;
      out.push({
        id: `capture.${c.kind}`,
        group: 'capture',
        text: t(c.text),
        examples: c.rest[lang].map((r) => `${prefix} ${r}`),
      });
    }
  }

  // Generated commands (one per settings section or tab) are summed up by
  // the example of "settings"; listing each would bury the rest.
  for (const spec of input.commands?.specs() ?? []) {
    if (spec.requires || spec.handler === 'tab' || !permitted(spec.role, viewer)) continue;
    const example = spec.terms[lang][0];
    if (example)
      out.push({
        id: `command.${spec.id}`,
        group: 'commands',
        text: input.commands?.title(spec) ?? spec.id,
        examples: [example],
      });
  }
  return out;
}
