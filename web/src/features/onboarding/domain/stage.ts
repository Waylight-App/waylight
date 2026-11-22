/** Which stage of the onboarding an account is in (#739). */
/** The stored step that ends the wizard; the tour follows. */
export const WIZARD_FINISHED = 'done';

export type Stage = 'wizard' | 'tour' | 'finished';

/** The wizard comes first, then the tour; a finished tour ends both. */
export function stageOf(status: { completed: boolean; step: string }): Stage {
  if (status.completed) return 'finished';
  return status.step === WIZARD_FINISHED ? 'tour' : 'wizard';
}
