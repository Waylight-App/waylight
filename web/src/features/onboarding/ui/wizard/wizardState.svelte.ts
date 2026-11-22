import type { Place } from '../../domain/profile';

/**
 * What the steps leave behind for each other. A class rather than a state
 * object of the shell, so the steps may write to it.
 */
export class WizardSession {
  /** The place saved in the profile step; the weather uses it. */
  place = $state<Place | undefined>();
  /** The choice made in the widget step, kept when going back. */
  selection = $state<{ template: string; kinds: string[] } | undefined>();
  /** The widget kinds created by the widget step. */
  created = $state<string[]>([]);
}

/** Set by a step: what Next does and what the footer says. */
export class StepControl {
  /** Runs before Next moves on; a thrown error stays on the step. */
  save: (() => Promise<void>) | undefined;
  /** Runs when the step is skipped. */
  skip: (() => void) | undefined;
  /** Short status shown in the footer, e.g. how many widgets are chosen. */
  status = $state('');
}
