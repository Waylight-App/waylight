import { describe, expect, it } from 'vitest';
import {
  cleanCredential,
  type Connection,
  type ConnectionService,
  groupByService,
  stateCounts,
  usageList,
  validateDraft,
} from './integrations';

const immich: ConnectionService = {
  service: 'immich',
  title: 'Immich',
  fields: [{ name: 'apiKey', secret: true, optional: false }],
};
const fuel: ConnectionService = {
  service: 'tankerkoenig',
  title: 'Tankerkönig',
  fixedUrl: 'https://creativecommons.tankerkoenig.de/json',
  fields: [{ name: 'apiKey', secret: true, optional: false }],
};

const conn = (id: string, service: string, name: string): Connection => ({
  id,
  service,
  name,
  url: 'http://x',
  fields: [],
  values: {},
  usedBy: [],
});

describe('integrations', () => {
  it('groups connections by service title and sorts them by name', () => {
    const groups = groupByService(
      [
        conn('1', 'tankerkoenig', 'Key'),
        conn('2', 'immich', 'Office'),
        conn('3', 'immich', 'Home'),
        conn('4', 'gone', 'Old'),
      ],
      [immich, fuel],
      'en',
    );
    expect(groups.map((g) => g.title)).toEqual(['gone', 'Immich', 'Tankerkönig']);
    expect(groups[1]?.connections.map((c) => c.name)).toEqual(['Home', 'Office']);
  });

  it('validates a draft against its service', () => {
    const ok = {
      service: 'immich',
      name: ' Home ',
      url: 'http://192.168.178.20:2283',
      credential: { apiKey: 'k' },
    };
    expect(validateDraft(ok, immich)).toBeNull();
    expect(validateDraft(ok, undefined)).toBe('integrations.error_service');
    expect(validateDraft({ ...ok, name: ' ' }, immich)).toBe('integrations.error_name');
    expect(validateDraft({ ...ok, name: 'x'.repeat(61) }, immich)).toBe('integrations.error_name_long');
    expect(validateDraft({ ...ok, url: '192.168.178.20' }, immich)).toBe('integrations.error_url');
    expect(validateDraft({ ...ok, credential: { apiKey: ' ' } }, immich)).toBe('integrations.error_field');
    // A service with one address needs none.
    expect(validateDraft({ ...ok, service: 'tankerkoenig', url: '' }, fuel)).toBeNull();
  });

  it('drops empty credential fields', () => {
    expect(cleanCredential({ username: ' anna ', password: '', fingerprint: '  ' })).toEqual({
      username: 'anna',
    });
  });

  it('names the widgets that use a connection, without other accounts’ tabs', () => {
    const list = usageList(
      [
        { widgetId: 'w1', tabId: 't1', kind: 'immich', mine: true },
        { widgetId: 'w2', tabId: 't9', kind: 'immich', mine: false },
        { widgetId: 'w3', tabId: 'gone', kind: 'immich', mine: true },
      ],
      () => 'Immich',
      (id) => (id === 't1' ? 'Home' : undefined),
      'another account',
    );
    expect(list).toEqual(['Immich (Home)', 'Immich (another account)', 'Immich']);
  });

  it('counts working and failing connections', () => {
    expect(
      stateCounts([{ kind: 'ok' }, { kind: 'failing', message: 'x' }, { kind: 'unchecked' }, { kind: 'ok' }]),
    ).toEqual({ ok: 2, failing: 1 });
  });
});
