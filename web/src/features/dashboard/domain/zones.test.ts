import { describe, expect, it } from 'vitest';
import type { WidgetCategory } from './catalog';
import type { Widget } from './dashboard';
import { defaultZone, groupByZone, parseZone, zoneOf } from './zones';

const widget = (id: string, kind: string, x: number, y: number, zone?: Widget['zone']): Widget => ({
  id,
  tabId: 't',
  kind,
  config: {},
  x,
  y,
  w: 3,
  h: 2,
  ...(zone ? { zone } : {}),
});

const categories: Record<string, WidgetCategory> = {
  clock: 'everyday',
  notes: 'everyday',
  services: 'server',
  adguard: 'network',
  rss: 'information',
};
const categoryOf = (kind: string) => categories[kind];

describe('zones', () => {
  it('parses known zones only', () => {
    expect(parseZone('system')).toBe('system');
    expect(parseZone('attic')).toBeUndefined();
    expect(parseZone(undefined)).toBeUndefined();
  });

  it('picks a default zone by kind and category', () => {
    expect(defaultZone('tonight', 'everyday')).toBe('today');
    expect(defaultZone('clock', 'everyday')).toBe('today');
    expect(defaultZone('todos', 'everyday')).toBe('today');
    expect(defaultZone('services', 'server')).toBe('system');
    expect(defaultZone('adguard', 'network')).toBe('system');
    expect(defaultZone('rss', 'information')).toBe('mine');
    expect(defaultZone('unknown', undefined)).toBe('mine');
  });

  it('prefers the zone a widget was moved to', () => {
    expect(zoneOf(widget('a', 'services', 0, 0, 'mine'), 'server')).toBe('mine');
  });

  it('groups in zone order and keeps the reading order inside a zone', () => {
    const groups = groupByZone(
      [
        widget('s', 'services', 0, 0),
        widget('n2', 'notes', 6, 4),
        widget('c', 'clock', 0, 2),
        widget('n1', 'notes', 0, 4),
      ],
      categoryOf,
      { system: true },
    );
    expect(groups.map((g) => [g.zone, g.widgets.map((w) => w.id)])).toEqual([
      ['today', ['c']],
      ['mine', ['n1', 'n2']],
      ['system', ['s']],
    ]);
  });

  it('hides the System zone from viewers who are not admins', () => {
    const groups = groupByZone([widget('s', 'services', 0, 0), widget('c', 'clock', 0, 1)], categoryOf, {
      system: false,
    });
    expect(groups.map((g) => g.zone)).toEqual(['today']);
  });

  it('keeps empty zones in edit mode', () => {
    const groups = groupByZone([], categoryOf, { system: true, keepEmpty: true });
    expect(groups.map((g) => g.zone)).toEqual(['today', 'mine', 'system']);
  });
});
