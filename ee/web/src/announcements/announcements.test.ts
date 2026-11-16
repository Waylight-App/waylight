import { describe, expect, it, vi } from 'vitest';
import { AnnouncementsApi, dayOf, endOfDay } from './announcements';

describe('announcements', () => {
  it('turns a day into the end of that local day and back', () => {
    const iso = endOfDay('2026-11-20');
    expect(iso).not.toBeNull();
    expect(dayOf(iso)).toBe('2026-11-20');
    expect(endOfDay('')).toBeNull();
    expect(dayOf(null)).toBe('');
  });

  it('creates or updates with a trimmed title and the expiry', async () => {
    const http = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    const api = new AnnouncementsApi(http);
    await api.save({
      title: ' Closed Friday ',
      body: 'See you',
      pinned: true,
      expiresOn: '',
      notify: false,
    });
    expect(http.post).toHaveBeenCalledWith('/api/announcements', {
      title: 'Closed Friday',
      body: 'See you',
      pinned: true,
      expiresAt: null,
      notify: false,
    });
    await api.save(
      {
        title: 'x',
        body: '',
        pinned: false,
        expiresOn: '2026-11-20',
        notify: true,
      },
      'a/1',
    );
    expect(http.put).toHaveBeenCalledWith(
      '/api/announcements/a%2F1',
      expect.objectContaining({ notify: true }),
    );
    await api.list(true);
    expect(http.get).toHaveBeenCalledWith('/api/announcements?archive=1');
  });

  it('asks the server whether the account may write', async () => {
    const http = {
      get: vi.fn().mockResolvedValue({ mayWrite: true }),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    expect(await new AnnouncementsApi(http).mayWrite()).toBe(true);
    expect(http.get).toHaveBeenCalledWith('/api/announcements/me');
  });
});
