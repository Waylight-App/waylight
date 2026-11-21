import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '../../../../shared/api-client/http';
import { CompactApi, toCompactState } from './compactApi';

describe('CompactApi (#595)', () => {
  it('reads the state from the module endpoint', async () => {
    const get = vi.fn().mockResolvedValue({
      title: { key: 'k1', text: '3 open', params: { count: '3' } },
      level: 'ok',
      count: 3,
      updatedAt: '2026-11-21T10:00:00Z',
    });
    const api = new CompactApi({ get } as unknown as HttpClient);
    const state = await api.state('todos');
    expect(get).toHaveBeenCalledWith('/api/todos/compact');
    expect(state).toEqual({
      title: { key: 'k1', text: '3 open', params: { count: '3' } },
      level: 'ok',
      count: 3,
      updatedAt: '2026-11-21T10:00:00Z',
    });
  });

  it('passes the query of a source that depends on the configuration (#614)', async () => {
    const get = vi
      .fn()
      .mockResolvedValue({ title: { key: 'k', text: 'Coffee' }, level: 'none', updatedAt: '' });
    await new CompactApi({ get } as unknown as HttpClient).state('smarthome?entity=switch.coffee');
    expect(get).toHaveBeenCalledWith('/api/smarthome/compact?entity=switch.coffee');
  });

  it('maps detail, an unknown level and an action with choices', () => {
    const state = toCompactState({
      title: { key: 'k', text: 'Office' },
      detail: { key: 'd', text: '8 in the office' },
      level: 'urgent',
      updatedAt: '2026-11-21T10:00:00Z',
      action: {
        id: 'presence.set',
        label: { key: 'l', text: 'Set status' },
        method: 'PUT',
        path: '/api/presence/me',
        choices: [
          { id: 'home', label: { key: 'h', text: 'Home' }, body: { status: 'home' }, selected: true },
        ],
      },
    });
    expect(state.level).toBe('none');
    expect(state.detail?.text).toBe('8 in the office');
    expect(state.title.params).toEqual({});
    expect(state.action?.choices).toEqual([
      { id: 'home', label: { key: 'h', text: 'Home', params: {} }, body: { status: 'home' }, selected: true },
    ]);
  });
});
