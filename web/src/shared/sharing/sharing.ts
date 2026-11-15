import type { HttpClient } from '../api-client/http';

/** A group a record can be shared with (ADR 0018, #301). */
export interface ShareGroup {
  id: string;
  name: string;
}

/** "private", "everyone" or "group:<id>". */
export type Visibility = string;

/** Visibility value of a group. */
export const groupVisibility = (id: string): Visibility => `group:${id}`;

/**
 * Loads the groups the signed-in account may share with: its own groups, or
 * every group for an admin. Used by the gateways of shareable features.
 */
export async function fetchShareGroups(http: HttpClient): Promise<ShareGroup[]> {
  const list = await http.get<{ id: string; name: string }[]>('/api/auth/my-groups');
  return list.map((g) => ({ id: g.id, name: g.name }));
}
