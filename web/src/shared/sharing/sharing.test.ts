import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '../api-client/http';
import { fetchShareGroups, groupVisibility, isUnknownGroup, sharedVia } from './sharing';

describe('sharing', () => {
  it('builds group visibilities and loads the groups', async () => {
    expect(groupVisibility('g1')).toBe('group:g1');
    const http = {
      get: vi.fn().mockResolvedValue([{ id: 'g1', name: 'Office', accounts: [] }]),
    } as unknown as HttpClient;
    await expect(fetchShareGroups(http)).resolves.toEqual([{ id: 'g1', name: 'Office' }]);
    expect(http.get).toHaveBeenCalledWith('/api/auth/my-groups');
  });

  it('spots a group value that is not in the list', () => {
    const groups = [{ id: 'g1', name: 'Office' }];
    expect(isUnknownGroup('group:g1', groups)).toBe(false);
    expect(isUnknownGroup('group:gone', groups)).toBe(true);
    expect(isUnknownGroup('private', groups)).toBe(false);
    expect(isUnknownGroup('everyone', [])).toBe(false);
  });
});

describe('sharedVia', () => {
  const groups = [{ id: 'g1', name: 'Office' }];

  it('names the group or everyone', () => {
    expect(sharedVia(groupVisibility('g1'), groups)).toEqual({ group: 'Office' });
    expect(sharedVia('everyone', groups)).toEqual({ everyone: true });
  });

  it('returns null for unknown groups and private records', () => {
    expect(sharedVia(groupVisibility('gone'), groups)).toBeNull();
    expect(sharedVia('private', groups)).toBeNull();
  });
});
