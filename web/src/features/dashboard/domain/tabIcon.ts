import type { IconName } from '../../../shared/ui/icons/paths';

/**
 * The line icon a tab gets in the bottom navigation of phones (#771). Tabs
 * are named by their owner, so the icon follows the name where it is a
 * familiar one (English or German) and falls back to a neutral one.
 */
const RULES: readonly [RegExp, IconName][] = [
  [/home|heim|zuhause|start/i, 'home'],
  [/server|system|lab|netz|network/i, 'server'],
  [/kalender|calendar|plan/i, 'calendar'],
  [/about|unterwegs|travel|reise/i, 'pin'],
];

export function tabIcon(name: string): IconName {
  return RULES.find(([re]) => re.test(name))?.[1] ?? 'bookmark';
}
