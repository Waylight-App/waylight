import { serverMessage } from '../../../shared/i18n';

/**
 * The short state of a widget (#597), as the server computes it for the
 * compact view, the app's small tiles and the watch. The web shows it in
 * small and strip tiles (#595).
 */
export type CompactLevel = 'none' | 'ok' | 'attention' | 'problem';

/** A text the server sends with its message key, so it can be translated. */
export interface CompactText {
  key: string;
  /** English text with the parameters filled in; the fallback. */
  text: string;
  params: Record<string, string>;
}

/** One chip of a compact action (presence: office, home, away). */
export interface CompactChoice {
  id: string;
  label: CompactText;
  body?: Record<string, unknown>;
  selected: boolean;
}

/** The one primary action of a compact view (#598): the request to send. */
export interface CompactAction {
  id: string;
  label: CompactText;
  method: string;
  path: string;
  body?: Record<string, unknown>;
  choices: CompactChoice[];
}

/** The short state of one widget. */
export interface CompactState {
  title: CompactText;
  detail?: CompactText;
  level: CompactLevel;
  count?: number;
  updatedAt: string;
  action?: CompactAction;
  /** Nothing to show; zoned dashboards hide the widget (#619). */
  empty?: boolean;
}

const LEVELS: readonly CompactLevel[] = ['none', 'ok', 'attention', 'problem'];

/** Parses a level from the API; a level from a newer server counts as none. */
export function parseLevel(value: string): CompactLevel {
  return LEVELS.find((l) => l === value) ?? 'none';
}

/** The text in the current language, falling back to the server's English text. */
export function compactText(t: CompactText): string {
  return serverMessage(t.key, t.params) ?? t.text;
}

/**
 * Whether a level gets a colour. Only what needs action is coloured, so a
 * dashboard of healthy widgets stays quiet (#593).
 */
export function isAlarming(level: CompactLevel): boolean {
  return level === 'attention' || level === 'problem';
}

/** How often compact states are read again. */
export const COMPACT_REFRESH_MS = 60_000;
