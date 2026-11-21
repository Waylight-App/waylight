import { describe, expect, it } from 'vitest';
import { byUrgency, type ListedProblem, listedEntries, sourceEntry, uniqueSources } from './attentionLine';

const source = { widgetId: 'w1', module: 'services', title: 'Services' };

describe('attention line (#593)', () => {
  it('lifts only amber and red states into the line', () => {
    expect(sourceEntry(source, { level: 'ok', text: '12 of 12 running' })).toBeNull();
    expect(sourceEntry(source, { level: 'none', text: 'No services yet' })).toBeNull();
    expect(sourceEntry(source, { level: 'problem', text: '11 of 12 running · Nextcloud is down' })).toEqual({
      id: 'compact:services',
      level: 'problem',
      title: 'Services',
      text: '11 of 12 running · Nextcloud is down',
      widgetId: 'w1',
    });
  });

  it('puts problems first and keeps the tab order within a level', () => {
    const e = (id: string, level: 'attention' | 'problem') => ({ id, level, title: id, text: '' });
    const ordered = byUrgency([e('a', 'attention'), e('b', 'problem'), e('c', 'attention'), e('d', 'problem')]);
    expect(ordered.map((x) => x.id)).toEqual(['b', 'd', 'a', 'c']);
  });

  it('takes failing connections of this tab or of no tab, not missing setup', () => {
    const items: ListedProblem[] = [
      { id: '1', kind: 'connection', title: 'Immich', text: 'Sign in again', widgetId: 'w9', tabId: 't1' },
      { id: '2', kind: 'connection', title: 'Mail', text: 'No answer', tabId: 't2' },
      { id: '3', kind: 'setup', title: 'Weather', text: 'Choose a place', tabId: 't1' },
      { id: '4', kind: 'connection', title: 'Calendar', text: 'Password refused' },
    ];
    expect(listedEntries(items, 't1')).toEqual([
      { id: 'attention:1', level: 'problem', title: 'Immich', text: 'Sign in again', widgetId: 'w9' },
      { id: 'attention:4', level: 'problem', title: 'Calendar', text: 'Password refused' },
    ]);
  });

  it('keeps one source per module', () => {
    const overview = { widgetId: 'w2', module: 'services', title: 'Status overview' };
    const backup = { widgetId: 'w3', module: 'backupmonitor', title: 'Backups' };
    expect(uniqueSources([source, overview, backup])).toEqual([source, backup]);
  });
});
