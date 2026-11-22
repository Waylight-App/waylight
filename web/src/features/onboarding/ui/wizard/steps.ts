import type { Component } from 'svelte';
import type { MessageKey } from '../../../../shared/i18n';
import type { StepDefinition, StepProps } from './wizardContext';
import DoneStep from './DoneStep.svelte';
import ProfileStep from './ProfileStep.svelte';
import SearchStep from './SearchStep.svelte';
import WelcomeStep from './WelcomeStep.svelte';
import WidgetsStep from './WidgetsStep.svelte';

/** A registered step: its layout, its component and the hint under bare steps. */
export interface RegisteredStep extends StepDefinition {
  component: Component<StepProps>;
  hint?: MessageKey;
}

/**
 * The wizard's steps as data; the shell does not know any of them. A new
 * step is one entry here and one component.
 */
export const stepRegistry: Record<'welcome' | 'profile' | 'widgets' | 'search' | 'done', RegisteredStep> = {
  welcome: {
    glow: 'middle',
    centered: true,
    footer: 'hint',
    component: WelcomeStep,
    hint: 'wizard.welcome.foot',
  },
  profile: { glow: 'top', centered: true, footer: 'actions', component: ProfileStep },
  widgets: { glow: 'top', centered: false, footer: 'actions', component: WidgetsStep },
  search: { glow: 'top', centered: false, footer: 'actions', component: SearchStep },
  done: { glow: 'middle', centered: true, footer: 'hint', component: DoneStep, hint: 'wizard.done.foot' },
};
