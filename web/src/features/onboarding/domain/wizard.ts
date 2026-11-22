import type { BuiltinTemplate } from '../../dashboard/domain/dashboard';

/**
 * The onboarding wizard (#739): one screen per step, before the dashboard.
 * Plugin steps are "extra:<id>".
 */
export type WizardStepId = 'welcome' | 'profile' | 'widgets' | 'done' | `extra:${string}`;

/** The stored step that ends the wizard; the tour follows. */
export const WIZARD_FINISHED = 'done';

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
  steps.push('done');
  return steps;
}

/** Where a stored step continues; an unknown or empty one starts at the beginning. */
export function startIndex(steps: readonly string[], stored: string): number {
  return Math.max(0, steps.indexOf(stored));
}

/** Which stage of the onboarding an account is in. */
export type Stage = 'wizard' | 'tour' | 'finished';

/** The wizard comes first, then the tour; a finished tour ends both. */
export function stageOf(status: { completed: boolean; step: string }): Stage {
  if (status.completed) return 'finished';
  return status.step === WIZARD_FINISHED ? 'tour' : 'wizard';
}

/** Widgets offered one by one, in this order; kinds that are switched off are left out. */
export const SINGLE_WIDGETS = ['calendar', 'shopping', 'todos', 'weather', 'waste', 'immich'] as const;

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
