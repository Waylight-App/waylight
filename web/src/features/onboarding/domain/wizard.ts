import type { BuiltinTemplate } from '../../dashboard/domain/dashboard';

/**
 * The onboarding wizard (#739): one screen per step, before the dashboard.
 * Plugin steps are "extra:<id>".
 */
export type WizardStepId = 'welcome' | 'profile' | 'widgets' | 'search' | 'done' | `extra:${string}`;

/**
 * The steps in order. The profile needs the personal settings; plugin steps
 * (the instance name) concern the whole installation, so owners only; the
 * widget choice only makes sense while the dashboard is still untouched.
 */
export function wizardSteps(o: {
  personal: boolean;
  owner: boolean;
  widgets: boolean;
  extra: readonly string[];
}): WizardStepId[] {
  const steps: WizardStepId[] = ['welcome'];
  if (o.personal) steps.push('profile');
  if (o.owner) steps.push(...o.extra.map((id) => `extra:${id}` as const));
  if (o.widgets) steps.push('widgets');
  steps.push('search', 'done');
  return steps;
}

/** Where a stored step continues; an unknown or empty one starts at the beginning. */
export function startIndex(steps: readonly string[], stored: string): number {
  return Math.max(0, steps.indexOf(stored));
}

/** Widgets offered one by one for the empty start page, in this order; kinds that are switched off are left out. */
export const SINGLE_WIDGETS = ['calendar', 'shopping', 'todos', 'weather', 'waste', 'immich'] as const;

/** The widget tiles of the step: those a template starts with and a few more that fit it. */
export interface Suggestions {
  selected: readonly string[];
  offered: readonly string[];
}

/**
 * The suggestions per built-in template. The preselected ones are the core of
 * the template; the offered ones are its other widgets and close relatives.
 * Anything else comes through the add-widget dialog.
 */
const SUGGESTIONS: Readonly<Record<string, Suggestions>> = {
  family: {
    selected: ['calendar', 'todos', 'shopping', 'meals'],
    offered: ['weather', 'waste', 'birthdays', 'chores'],
  },
  homelab: {
    selected: ['clock', 'todos', 'bookmarks', 'services'],
    offered: ['docker', 'backup-status', 'diskhealth', 'grafana-alerts'],
  },
  work: {
    selected: ['calendar', 'todos', 'focus', 'notes'],
    offered: ['mail', 'bookmarks', 'services', 'readlater'],
  },
};

/** Without a template (the empty start page) nothing is preselected and the single widgets are offered. */
const NO_TEMPLATE: Suggestions = { selected: [], offered: SINGLE_WIDGETS };

/** The suggestions of a template id; unknown ids and the empty template get the single widgets. */
export function suggestionsOf(template: string): Suggestions {
  return SUGGESTIONS[template] ?? NO_TEMPLATE;
}

/**
 * The tiles to draw: the suggestions that are available, then any other
 * selected kind (chosen in the add-widget dialog) so every choice stays visible.
 */
export function suggestionTiles(
  s: Suggestions,
  chosen: readonly string[],
  available: ReadonlySet<string>,
): string[] {
  const own = [...s.selected, ...s.offered].filter((k) => available.has(k));
  return [...own, ...chosen.filter((k) => !own.includes(k))];
}

/** Built-in templates shown as cards, in this order. */
const TEMPLATE_CARDS = ['family', 'homelab', 'work'] as const;

/** The built-in templates the wizard offers. */
export function templateCards(list: readonly BuiltinTemplate[]): BuiltinTemplate[] {
  return TEMPLATE_CARDS.flatMap((id) => list.filter((t) => t.id === id));
}

/** The widget kinds of a template, each once, in layout order. */
export function templateKinds(t: Pick<BuiltinTemplate, 'widgets'>): string[] {
  return [...new Set(t.widgets.map((w) => w.kind))];
}

/** What a created widget still needs before it shows data. */
export type LaterSetup = 'set_up' | 'connect';

/**
 * Widgets that cannot work without the user: a waste calendar address, or an
 * account of an external service. They are set up on the dashboard.
 */
const LATER_SETUP: Readonly<Record<string, LaterSetup>> = {
  waste: 'set_up',
  immich: 'connect',
  mail: 'connect',
  docker: 'connect',
  media: 'connect',
  smarthome: 'connect',
  proxmox: 'connect',
  transit: 'set_up',
};

/** The weather needs a place; without one it is set up on the dashboard. */
export function laterSetup(kind: string, hasPlace: boolean): LaterSetup | undefined {
  if (kind === 'weather' && !hasPlace) return 'set_up';
  return LATER_SETUP[kind];
}

/** A query offered in the search step, with the module it needs, if any. */
export interface SearchExample {
  query: string;
  module?: string;
}

/** The example sets: they follow the template chosen in the widget step. */
export type ExampleSet = 'family' | 'home' | 'homelab';

/** The set for a chosen template; the empty template and single widgets get the home set. */
export function exampleSetOf(template: string): ExampleSet {
  return template === 'family' || template === 'homelab' ? template : 'home';
}

/**
 * The queries offered to try, per language and set, in display order; the
 * first one is typed by the animation. Each is answered by the real search
 * (a test checks this), and shown only when its module is on. To change the
 * wording, edit the lists; nothing else refers to them.
 */
export const SEARCH_EXAMPLES: Record<'de' | 'en', Record<ExampleSet, readonly SearchExample[]>> = {
  de: {
    family: [
      { query: 'Was steht heute an?', module: 'calendar' },
      { query: 'Wer hat bald Geburtstag?', module: 'birthdays' },
      { query: 'Wann kommt die Biotonne?', module: 'waste' },
      { query: 'Brauche ich morgen einen Regenschirm?', module: 'weather' },
      { query: 'Was steht auf der Einkaufsliste?', module: 'shopping' },
    ],
    home: [
      { query: 'Wann kommt die Biotonne?', module: 'waste' },
      { query: 'Wie warm wird es am Wochenende?', module: 'weather' },
      { query: 'Was steht auf der Einkaufsliste?', module: 'shopping' },
      { query: 'Was steht heute an?', module: 'calendar' },
      { query: 'Wann sind die nächsten Ferien?', module: 'holidays' },
    ],
    homelab: [
      { query: 'Läuft alles bei meinem Server?', module: 'services' },
      // Only typed in: the wizard never runs an action.
      { query: 'Starte Jellyfin neu', module: 'docker' },
      { query: 'Wie viel Speicher ist frei?', module: 'system' },
      { query: 'Was steht heute an?', module: 'calendar' },
    ],
  },
  en: {
    family: [
      { query: 'What is on today?', module: 'calendar' },
      { query: 'Who has a birthday soon?', module: 'birthdays' },
      { query: 'When is the organic waste collected?', module: 'waste' },
      { query: 'Do I need an umbrella tomorrow?', module: 'weather' },
      { query: 'What is on the shopping list?', module: 'shopping' },
    ],
    home: [
      { query: 'When is the organic waste collected?', module: 'waste' },
      { query: 'Will it rain on the weekend?', module: 'weather' },
      { query: 'What is on the shopping list?', module: 'shopping' },
      { query: 'What is on today?', module: 'calendar' },
      { query: 'When are the next holidays?', module: 'holidays' },
    ],
    homelab: [
      { query: 'Is everything running on my server?', module: 'services' },
      { query: 'Restart Jellyfin', module: 'docker' },
      { query: 'How much storage is free?', module: 'system' },
      { query: 'What is on today?', module: 'calendar' },
    ],
  },
};

/** The calculator needs no module; it fills the list up when few modules are on. */
const FALLBACK = { de: ['19 % von 240', '5 km in mi'], en: ['19 % of 240', '5 km in mi'] } as const;

/** The examples of a language and set whose module is on, at least two. */
export function searchExamples(lang: 'de' | 'en', set: ExampleSet, modules: ReadonlySet<string>): string[] {
  const own = SEARCH_EXAMPLES[lang][set]
    .filter((e) => e.module === undefined || modules.has(e.module))
    .map((e) => e.query);
  return own.length >= 2 ? own : [...own, ...FALLBACK[lang]].slice(0, 3);
}
