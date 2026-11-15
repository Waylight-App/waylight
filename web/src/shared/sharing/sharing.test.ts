import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '../api-client/http';
import { fetchShareGroups, groupVisibility } from './sharing';

describe('sharing', () => {
  it('builds group visibilities and loads the groups', async () => {
    expect(groupVisibility('g1')).toBe('group:g1');
    const http = {
      get: vi.fn().mockResolvedValue([{ id: 'g1', name: 'Office', accounts: [] }]),
    } as unknown as HttpClient;
    await expect(fetchShareGroups(http)).resolves.toEqual([{ id: 'g1', name: 'Office' }]);
    expect(http.get).toHaveBeenCalledWith('/api/auth/my-groups');
  });
});
