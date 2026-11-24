import type { IconName } from '../../../shared/ui/icons/paths';

/**
 * The line icon a tab gets in the bottom navigation of phones (#771). Tabs
 * are named by their owner, so the icon follows the name where it is a
 * familiar one (English or German) and falls back to a neutral one.
 */
const RULES: readonly [RegExp, IconName][] = [
  [/home|start|heim|zuhause|overview|übersicht/i, 'home'],
  [/server|system|lab|infra|netz|network|admin|docker/i, 'server'],
  [/calendar|kalender|plan|termin/i, 'calendar'],
  [/famil|haushalt|household|kids|kinder/i, 'users'],
  [/news|feed|rss|read|lesen/i, 'rss'],
  [/out|about|unterwegs|travel|reise|trip/i, 'pin'],
  [/note|notiz|wiki|doc|work|arbeit|büro|office/i, 'note'],
];

export function tabIcon(name: string): IconName {
  return RULES.find(([re]) => re.test(name))?.[1] ?? 'bookmark';
}
