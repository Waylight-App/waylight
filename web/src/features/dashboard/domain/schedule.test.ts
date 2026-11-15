import { describe, expect, it } from 'vitest';
import type { Tab, TabRule } from './dashboard';
import { defaultTabId, ruleMatches } from './schedule';

const rule = (r: Partial<TabRule>): TabRule => ({ days: [], from: '', to: '', networks: [], ...r });
// 2026-11-16 is a Monday.
const at = (hhmm: string, day = 16) => new Date(`2026-11-${day}T${hhmm}:00`);

describe('ruleMatches (mirrors schedule.go)', () => {
  it('matches all day without times and days', () => {
    expect(ruleMatches(rule({}), at('03:00'))).toBe(true);
  });
  it('matches inside a time range, not at its end', () => {
    const r = rule({ from: '08:00', to: '17:00' });
    expect(ruleMatches(r, at('08:00'))).toBe(true);
    expect(ruleMatches(r, at('16:59'))).toBe(true);
    expect(ruleMatches(r, at('17:00'))).toBe(false);
  });
  it('gives the early hours of a range over midnight to the day it started', () => {
    const fridayNight = rule({ days: [5], from: '22:00', to: '02:00' });
    expect(ruleMatches(fridayNight, at('23:00', 20))).toBe(true); // Friday
    expect(ruleMatches(fridayNight, at('01:00', 21))).toBe(true); // Saturday morning
    expect(ruleMatches(fridayNight, at('01:00', 20))).toBe(false); // Friday morning = Thursday's night
    expect(ruleMatches(fridayNight, at('12:00', 20))).toBe(false);
  });
  it('checks the weekday', () => {
    expect(ruleMatches(rule({ days: [1] }), at('12:00'))).toBe(true);
    expect(ruleMatches(rule({ days: [0, 6] }), at('12:00'))).toBe(false);
  });
  it('never matches a rule with networks, since the browser cannot tell the network', () => {
    expect(ruleMatches(rule({ networks: ['192.168.0.0/24'] }), at('12:00'))).toBe(false);
  });
});

describe('defaultTabId', () => {
  const tab = (id: string, rules: TabRule[]): Tab => ({
    id,
    name: id,
    position: 0,
    background: '',
    rules,
    widgets: [],
    mobile: [],
    mobileCustom: false,
  });
  it('returns the first tab with a matching rule, or ""', () => {
    const tabs = [tab('a', []), tab('b', [rule({ days: [1] })]), tab('c', [rule({})])];
    expect(defaultTabId(tabs, at('12:00'))).toBe('b');
    expect(defaultTabId([tab('a', [])], at('12:00'))).toBe('');
  });
});
