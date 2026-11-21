import type { KioskInput } from '../../features/auth/domain/auth';

/**
 * The wall tablet assistant (#680) as data: the choices of its one form and
 * the kiosk display they become. Defaults suit a family screen in a hall or
 * kitchen: large text, dimmed at night, the "Wall tablet" template.
 */
export interface WallTabletChoice {
  name: string;
  /** "template" creates a new tab from the Wall tablet template. */
  source: 'template' | 'tabs';
  /** The existing tabs, when source is "tabs". */
  tabs: string[];
  largeText: boolean;
  night: boolean;
  nightFrom: string;
  nightTo: string;
  /** Seconds per tab; only used with more than one tab. */
  rotateSeconds: number;
}

/** The id of the built-in template the assistant offers. */
export const WALL_TEMPLATE = 'wall';

/** Seconds per tab the assistant suggests once several tabs are chosen. */
export const DEFAULT_ROTATION = 60;

export function defaultChoice(): WallTabletChoice {
  return {
    name: 'wall-tablet',
    source: 'template',
    tabs: [],
    largeText: true,
    night: true,
    nightFrom: '22:00',
    nightTo: '06:30',
    rotateSeconds: DEFAULT_ROTATION,
  };
}

/** Whether the form can be sent: a name and at least one tab. */
export function ready(c: WallTabletChoice): boolean {
  return c.name.trim() !== '' && (c.source === 'template' || c.tabs.length > 0);
}

/**
 * The display to create; templateTab is the tab the template just created.
 * One tab never rotates, so the rotation only applies to several.
 */
export function kioskInput(c: WallTabletChoice, templateTab?: string): KioskInput {
  const tabs = c.source === 'template' && templateTab ? [templateTab] : c.tabs;
  return {
    name: c.name.trim(),
    tabs,
    disabled: false,
    rotateSeconds: tabs.length > 1 ? c.rotateSeconds : 0,
    nightFrom: c.night ? c.nightFrom : '',
    nightTo: c.night ? c.nightTo : '',
    largeText: c.largeText,
  };
}
