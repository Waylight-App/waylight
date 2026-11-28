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
 * True when a value names a group that is not in the list: the group was
 * deleted, or the account is no longer in it. The record is then visible to
 * its owner only, and the choice must still show that state (#390).
 */
export const isUnknownGroup = (value: Visibility, groups: ShareGroup[]): boolean =>
  value.startsWith('group:') && !groups.some((g) => groupVisibility(g.id) === value);

/**
 * Through what a record is shared with the viewer (#420): everyone, a
 * named group, or null when the group is not known (or the record private).
 */
export function sharedVia(
  value: Visibility,
  groups: ShareGroup[],
): { everyone: true } | { group: string } | null {
  if (value === 'everyone') return { everyone: true };
  const g = groups.find((x) => groupVisibility(x.id) === value);
  return g ? { group: g.name } : null;
}

/**
 * Loads the groups the signed-in account may share with: its own groups, or
 * every group for an admin. Used by the gateways of shareable features.
 */
export async function fetchShareGroups(http: HttpClient): Promise<ShareGroup[]> {
  const list = await http.get<{ id: string; name: string }[]>('/api/auth/my-groups');
  return list.map((g) => ({ id: g.id, name: g.name }));
}

/**
 * What a widget says about who shares a record (#872): a group by name,
 * everyone, a group the viewer does not know (an admin's share), or null
 * for a private record, which needs no mark.
 */
export function sharedState(
  value: Visibility,
  groups: ShareGroup[],
): { everyone: true } | { group: string } | { unknownGroup: true } | null {
  if (!value || value === 'private') return null;
  return sharedVia(value, groups) ?? { unknownGroup: true };
}
