import { stageOf, WIZARD_FINISHED, type Stage } from '../domain/wizard';

/** The stored onboarding state of the account. */
export interface OnboardingStatus {
  completed: boolean;
  /** The wizard step the account stopped at; empty before it started. */
  step: string;
}

/** Output port: onboarding backend. */
export interface OnboardingGateway {
  status(): Promise<OnboardingStatus>;
  setCompleted(completed: boolean): Promise<void>;
  /** Stores the wizard step and makes the tour pending. */
  setStep(step: string): Promise<void>;
}

/** Application logic of the onboarding: the wizard, then the guided tour. */
export class OnboardingUseCases {
  constructor(private readonly gateway: OnboardingGateway) {}

  /** What to show now, and the wizard step to continue at. */
  async state(): Promise<{ stage: Stage; step: string }> {
    const status = await this.gateway.status();
    return { stage: stageOf(status), step: status.step };
  }

  /** Remembers the wizard step, so a reload continues there. */
  saveStep(step: string): Promise<void> {
    return this.gateway.setStep(step);
  }

  /** The wizard is finished or skipped; the tour is next. */
  finishWizard(): Promise<void> {
    return this.gateway.setStep(WIZARD_FINISHED);
  }

  /** Marks the tour as done (finished or skipped). */
  finish(): Promise<void> {
    return this.gateway.setCompleted(true);
  }

  /** Makes the tour appear again, without the wizard. */
  restart(): Promise<void> {
    return this.gateway.setStep(WIZARD_FINISHED);
  }
}
