import { describe, expect, it, vi } from 'vitest';
import { TeamCalendarApi } from './teamcalendar';

describe('TeamCalendarApi', () => {
  it('shares a subscription and loads the own groups', async () => {
    const http = {
      get: vi.fn().mockResolvedValue([{ id: 'g1', name: 'Office', accounts: [] }]),
      post: vi.fn(),
      put: vi.fn().mockResolvedValue({ visibility: 'group:g1' }),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    const api = new TeamCalendarApi(http);
    expect(await api.groups()).toEqual([{ id: 'g1', name: 'Office' }]);
    expect(http.get).toHaveBeenCalledWith('/api/auth/my-groups');
    await api.share('s/1', 'group:g1');
    expect(http.put).toHaveBeenCalledWith('/api/teamcalendar/subscriptions/s%2F1', {
      visibility: 'group:g1',
    });
  });

  it('shares a connected account on its own route', async () => {
    const http = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn().mockResolvedValue({ visibility: 'everyone' }),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    await new TeamCalendarApi(http).shareAccount('a1', 'everyone');
    expect(http.put).toHaveBeenCalledWith('/api/teamcalendar/accounts/a1', { visibility: 'everyone' });
  });
});
