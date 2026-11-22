import type { Snippet } from 'svelte';
import type { DashboardUseCases } from '../../../dashboard/usecase/dashboardUseCases';
import type { WidgetType } from '../../../dashboard/domain/dashboard';
import type { IconName } from '../../../../shared/ui/icons/paths';
import type { Answer } from '../../../search/domain/smart/answer';
import type { Personal, Place } from '../../domain/profile';
import type { WizardStepId } from '../../domain/wizard';
import type { StepControl, WizardSession } from './wizardState.svelte';

/** What the wizard may use of the app; the shell wires the real ports in. */
export interface WizardContext {
  /** Invited members get the personal steps only. */
  owner: boolean;
  /** The personal settings as they are now; undefined hides the profile step. */
  personal?: Personal;
  /** Saves changed personal settings. */
  savePersonal: (patch: Partial<Personal>) => Promise<void>;
  /** Shows a change right away (the language); null undoes the preview. */
  preview: (patch: Partial<Personal> | null) => void;
  /** Saved places (#591); undefined hides the place field. */
  places?: {
    /** Place search (the weather geocoder); omitted when it is off. */
    search?: (query: string) => Promise<Omit<Place, 'id'>[]>;
    save: (place: Omit<Place, 'id'>) => Promise<Place>;
  };
  /** The widget choice; undefined where the dashboard is already set up. */
  widgets?: {
    uc: DashboardUseCases;
    /** The widget types that are switched on, with their icons. */
    types: { type: WidgetType; icon: IconName }[];
    /** The example tabs of the setup are still there (#670). */
    sample: boolean;
  };
  /** Plugin steps, e.g. the instance name (owner only). */
  extraSteps: { id: string; title: string }[];
  /** Draws a plugin step; it registers what Next saves. */
  extra?: Snippet<[string, (save: () => Promise<void>) => void]>;
  /** The search of the app, for the try-it step; its answers are the real ones. */
  search: {
    /** Example queries that work for this account, in the current language, for the chosen template. */
    examples: (template: string) => Promise<string[]>;
    /** The answers of the smart search; the saved place serves the weather. */
    answer: (query: string, signal: AbortSignal, place?: Place) => Promise<Answer[]>;
  };
  /** Title of a widget kind. */
  widgetName: (kind: string) => string;
  /** Icon of a widget kind. */
  widgetIcon: (kind: string) => IconName | undefined;
  /** What a widget needs to be set up and where the documentation explains it. */
  setupHelp: (kind: string) => Promise<{ needs: string; docsUrl: string } | undefined>;
  /** Hands the widgets that still need setting up to the dashboard, which leads through them. */
  startSetup: (kinds: string[]) => void;
}

/** What a step can ask the shell to do. */
export interface StepActions {
  next: () => void;
  /** Ends the wizard without the remaining steps. */
  skipAll: () => void;
}

/** Props every step component gets. */
export interface StepProps {
  ctx: WizardContext;
  session: WizardSession;
  ctl: StepControl;
  actions: StepActions;
}

/** A step as data: what the shell needs to lay it out. */
export interface StepDefinition {
  /** Where the warm glow sits: the middle of the screen or its top edge. */
  glow: 'middle' | 'top';
  /** The content is centred vertically, else it starts below the header. */
  centered: boolean;
  /** Bare steps have their own buttons; the footer then only holds a hint. */
  footer: 'actions' | 'hint';
}

export type StepId = WizardStepId;
