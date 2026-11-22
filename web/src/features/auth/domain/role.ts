import type { MessageKey } from '../../../shared/i18n';

/** The four account types; the technical names stay in the API and in config-as-code. */
export type Role = 'owner' | 'admin' | 'member' | 'guest';

/** Plain-words label and one-sentence explanation of each role (#679). */
export const roleTexts: Record<Role, { label: MessageKey; hint: MessageKey }> = {
  owner: { label: 'auth.role.owner', hint: 'auth.role.owner_hint' },
  admin: { label: 'auth.role.admin', hint: 'auth.role.admin_hint' },
  member: { label: 'auth.role.member', hint: 'auth.role.member_hint' },
  guest: { label: 'auth.role.guest', hint: 'auth.role.guest_hint' },
};
