import { t } from '../../../shared/i18n';
/** Minimum password length; mirrors the backend password policy. */
export const MIN_PASSWORD_LENGTH = 12;

/** Authentication status of the current browser. */
export interface AuthStatus {
  setupRequired: boolean;
  authenticated: boolean;
  csrfToken: string | null;
  /** Public demo instance: signs in without a password. */
  demo: boolean;
  /** Read-only guest account (ADR 0012); also true for kiosk displays. */
  guest: boolean;
  /** Kiosk display (#50): no controls, rotation, night dimming. */
  kiosk: boolean;
  /** Admin of the instance: the owner or an admin (ADR 0022). */
  admin: boolean;
  /** The owner: backups, export and API tokens stay with this account (ADR 0022). */
  owner: boolean;
  /** Name of the signed-in account; '' when signed out. */
  username: string;
  /** Member: own tabs and content, no instance settings (#294). */
  member: boolean;
  /** Available sign-in methods (ADR 0013). */
  sso: { oidc: boolean; header: boolean; localLogin: boolean };
  /** Members and guests can ask for a reset link by e-mail (#296). */
  passwordReset: boolean;
}

/** Result of a single sign-on attempt. */
export type SsoResult = { kind: 'signed-in' } | { kind: 'needs-code'; ticket: string };

/** Message key for an error the provider redirect reported (?sso_error=…). */
export function ssoErrorKey(
  code: string | null,
): 'auth.sso_error.not_allowed' | 'auth.sso_error.failed' | null {
  if (!code) return null;
  return code === 'not_allowed' ? 'auth.sso_error.not_allowed' : 'auth.sso_error.failed';
}

/** A read-only guest account as managed by the owner. */
export interface Guest {
  id: string;
  name: string;
  tabs: string[];
  disabled: boolean;
  createdAt: Date;
}

/** A member account as managed by an admin (#294). */
export interface Member {
  id: string;
  name: string;
  groups: string[];
  disabled: boolean;
  createdAt: Date;
  /** "member" or "admin" (ADR 0022). */
  role: 'member' | 'admin';
  /** Where an admin role comes from: '' (set by hand), "sso" or "directory". */
  roleSource: string;
}

/** The members and whether new ones can be created in this version. */
export interface MemberList {
  available: boolean;
  members: Member[];
}

/** A group that records can be shared with (ADR 0018). */
export interface Group {
  id: string;
  name: string;
  /** IDs of the members and guests in the group. */
  accounts: string[];
}

/** Guest names: 2–32 lowercase letters, digits, dots, dashes, underscores. */
export function validGuestName(name: string): boolean {
  return /^[a-z0-9][a-z0-9._-]{1,31}$/.test(name) && name !== 'owner';
}

/** A freshly started session. */
export interface Session {
  csrfToken: string;
  expiresAt: Date;
}

/** Which screen the app has to show. */
export type AuthView = 'setup' | 'login' | 'app';

/** Derives the screen to show from the authentication status. */
export function viewFor(status: AuthStatus): AuthView {
  if (status.setupRequired) return 'setup';
  return status.authenticated ? 'app' : 'login';
}

/**
 * Validates a new password and its confirmation.
 * @returns an error message, or null if the password is acceptable
 */
export function validateNewPassword(password: string, confirmation: string): string | null {
  if ([...password].length < MIN_PASSWORD_LENGTH) {
    return t('auth.password_too_short', { min: MIN_PASSWORD_LENGTH });
  }
  if (password.trim() === '') return t('auth.password_whitespace');
  if (password !== confirmation) return t('auth.passwords_differ');
  return null;
}

/** How a kiosk display shows its tabs (#50). */
export interface KioskOptions {
  /** Seconds per tab; 0 = no rotation. */
  rotateSeconds: number;
  /** "HH:MM" local time; both empty = never dimmed. */
  nightFrom: string;
  nightTo: string;
}

/** A kiosk display as managed by the owner. */
export interface KioskDisplay extends KioskOptions {
  id: string;
  name: string;
  tabs: string[];
  disabled: boolean;
  createdAt: Date;
  lastUsedAt: Date | null;
}

/** Input of creating or changing a display. */
export interface KioskInput extends KioskOptions {
  name: string;
  tabs: string[];
  disabled: boolean;
}

/** An open invitation (#296). */
export interface Invite {
  id: string;
  name: string;
  role: 'member' | 'guest';
  tabs: string[];
  createdAt: Date;
  expiresAt: Date;
}

/** Input of an invitation. */
export interface InviteInput {
  name: string;
  role: 'member' | 'guest';
  tabs: string[];
  validDays: number;
}

/** What an invitation or reset link is for. */
export interface LinkInfo {
  purpose: 'invite' | 'reset';
  name: string;
}

/** The page behind invitation and reset links; the token stays in the fragment. */
export const accountLinkPath = '/account-link';

/** The link of an invitation or reset token. */
export function accountLink(origin: string, token: string): string {
  return `${origin}${accountLinkPath}#${encodeURIComponent(token)}`;
}

/** The link that signs a display in; the token travels in the fragment, never to the server log. */
export function kioskLink(origin: string, token: string): string {
  return `${origin}/kiosk#${encodeURIComponent(token)}`;
}

/** Whether a "HH:MM" time window (possibly over midnight) contains now. */
export function inNightWindow(now: Date, from: string, to: string): boolean {
  if (!from || !to) return false;
  const minutes = (s: string) => Number(s.slice(0, 2)) * 60 + Number(s.slice(3, 5));
  const m = now.getHours() * 60 + now.getMinutes();
  const a = minutes(from);
  const b = minutes(to);
  return a < b ? m >= a && m < b : m >= a || m < b;
}

/** Remaining validity of an invitation (#421): whole days, hours on the last day. */
export type Remaining = { unit: 'days' | 'hours'; n: number } | { unit: 'expired' };

/** How long something expiring at `expires` is still valid at `now`. */
export function remainingTime(expires: Date, now: Date): Remaining {
  const ms = expires.getTime() - now.getTime();
  if (ms <= 0) return { unit: 'expired' };
  const hours = ms / 3_600_000;
  if (hours < 24) return { unit: 'hours', n: Math.max(1, Math.ceil(hours)) };
  return { unit: 'days', n: Math.floor(hours / 24) };
}
