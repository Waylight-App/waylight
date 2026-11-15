import type { Tab, TabRule } from './dashboard';

const minutes = (hhmm: string): number => {
  const [h = '0', m = '0'] = hhmm.split(':');
  return Number(h) * 60 + Number(m);
};

/**
 * Whether a default-tab rule (#71) applies at `now`. Mirrors TabRule.Matches
 * of the server (internal/modules/dashboard/domain/schedule.go), which the
 * browser extension's local mode cannot ask (ADR 0019). A range over
 * midnight (22:00–02:00) belongs to the weekday it starts on.
 *
 * The browser does not know which network it is in, so a rule with networks
 * never matches here.
 */
export function ruleMatches(rule: TabRule, now: Date): boolean {
  if (rule.networks.length > 0) return false;
  const m = now.getHours() * 60 + now.getMinutes();
  let day = now.getDay();
  if (rule.from !== rule.to) {
    const from = minutes(rule.from);
    const to = minutes(rule.to);
    if (from < to) {
      if (m < from || m >= to) return false;
    } else if (m < to) {
      day = (day + 6) % 7; // the early hours belong to the previous day's rule
    } else if (m < from) {
      return false;
    }
  }
  return rule.days.length === 0 || rule.days.includes(day);
}

/** The first tab (in tab order) with a matching rule, or "" to keep the last used tab. */
export function defaultTabId(tabs: readonly Tab[], now: Date): string {
  return tabs.find((t) => t.rules.some((r) => ruleMatches(r, now)))?.id ?? '';
}
