import { describe, expect, it } from 'vitest';
import {
  actionTarget,
  type AttentionItem,
  connectionCounts,
  groupItems,
  sameItems,
  type ServiceConnection,
  sortConnections,
} from './attention';

const item = (over: Partial<AttentionItem>): AttentionItem => ({
  id: 'module:immich',
  kind: 'connection',
  integration: 'immich',
  reason: { message: 'Immich does not answer.', messageKey: 'srv_1' },
  since: new Date('2026-11-21T10:00:00Z'),
  action: 'check',
  ...over,
});

const sections: Record<string, string> = { immich: 'immich' };
const sectionOf = (i: string) => sections[i];

describe('groupItems', () => {
  it('puts failures first and keeps the order within each group', () => {
    const a = item({ id: 'a', kind: 'setup' });
    const b = item({ id: 'b' });
    const c = item({ id: 'c', kind: 'setup' });
    expect(groupItems([a, b, c])).toEqual({ failing: [b], setup: [a, c] });
  });
});

describe('actionTarget', () => {
  it('checks again in place', () => {
    expect(actionTarget(item({ action: 'check' }), sectionOf)).toEqual({ kind: 'check' });
  });

  it('opens the widget for a missing widget setting', () => {
    const it = item({ kind: 'setup', action: 'choose', integration: 'transit', widgetId: 'w1', tabId: 't1' });
    expect(actionTarget(it, sectionOf)).toEqual({ kind: 'widget', tabId: 't1', widgetId: 'w1' });
    expect(actionTarget({ ...it, widgetId: undefined }, sectionOf)).toEqual({ kind: 'none' });
  });

  it('opens the integration settings to connect or fix', () => {
    for (const action of ['connect', 'fix', 'signIn'] as const)
      expect(actionTarget(item({ action, widgetId: 'w1', tabId: 't1' }), sectionOf)).toEqual({
        kind: 'settings',
        section: 'immich',
      });
  });

  it('falls back to the widget when the integration has no settings section', () => {
    const it = item({ action: 'connect', integration: 'other', widgetId: 'w1', tabId: 't1' });
    expect(actionTarget(it, sectionOf)).toEqual({ kind: 'widget', tabId: 't1', widgetId: 'w1' });
  });
});

describe('sameItems', () => {
  it('compares id, action and reason', () => {
    const a = [item({})];
    expect(sameItems(a, [item({})])).toBe(true);
    expect(sameItems(a, [item({ action: 'fix' })])).toBe(false);
    expect(sameItems(a, [])).toBe(false);
  });
});

describe('connections', () => {
  const conn = (integration: string, state: ServiceConnection['state']): ServiceConnection => ({
    integration,
    instance: true,
    state,
    widgets: [],
  });
  const list = [
    conn('proxmox', 'setup'),
    conn('immich', 'ok'),
    conn('grafana', 'failing'),
    conn('arr', 'ok'),
  ];

  it('puts problems first, then connected, then the rest, by name', () => {
    expect(sortConnections(list, (i) => i).map((c) => c.integration)).toEqual([
      'grafana',
      'arr',
      'immich',
      'proxmox',
    ]);
  });

  it('counts for the summary', () => {
    expect(connectionCounts(list)).toEqual({ connected: 2, failing: 1, setup: 1 });
  });
});
