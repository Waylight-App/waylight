import type { LinkedSignIn, MicrosoftFlow } from '../domain/identity';
import type { ApiToken } from '../domain/apiToken';
import type { PairingCode } from '../domain/appPairing';
import type {
  Passkey,
  PasskeyAssertion,
  PasskeyAttestation,
  PasskeyRegistrationOptions,
  SecondFactorChallenge,
} from '../domain/passkey';
import type {
  AuthStatus,
  Group,
  Guest,
  Invite,
  InviteInput,
  KioskDisplay,
  KioskInput,
  KioskOptions,
  LinkInfo,
  MemberList,
  Member,
  Session,
} from '../domain/auth';
import type { SecurityEvent } from '../domain/securityEvent';
import type { SessionInfo } from '../domain/session';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../domain/twoFactor';

/** Output port: talks to the authentication backend. */
export interface AuthGateway {
  status(): Promise<AuthStatus>;
  /** profile: what the start page shows (#281). */
  setup(setupToken: string, password: string, profile?: 'everyday' | 'homelab'): Promise<Session>;
  /** Returns a challenge instead of a session when a passkey is registered. */
  login(password: string, code?: string, username?: string): Promise<Session | SecondFactorChallenge>;
  passkeyLogin(ticket: string, assertion: PasskeyAssertion): Promise<Session>;
  passkeys(): Promise<{ available: boolean; passkeys: Passkey[] }>;
  beginPasskeyRegistration(): Promise<PasskeyRegistrationOptions>;
  finishPasskeyRegistration(name: string, attestation: PasskeyAttestation): Promise<string[]>;
  deletePasskey(id: string): Promise<void>;
  /** Sign-ins at the OpenID provider linked to this account (#527). */
  linkedSignIns(): Promise<LinkedSignIn[]>;
  /** Starts linking; returns the provider URL to open. */
  beginLinkSignIn(): Promise<string>;
  unlinkSignIn(subject: string): Promise<void>;
  /** Starts a Microsoft code flow: to link (signed in) or to sign in (#535). */
  startMicrosoft(link: boolean): Promise<MicrosoftFlow>;
  /** Polls it; a sign-in answers with the session. */
  pollMicrosoft(flow: string): Promise<{ pending: true } | { linked: true } | { ticket: string } | Session>;
  /** Accepts an invitation through the provider (#528); returns its URL. */
  signUpWithSSO(token: string): Promise<string>;
  guests(): Promise<Guest[]>;
  createGuest(name: string, password: string, tabs: string[]): Promise<Guest>;
  updateGuest(id: string, tabs: string[], disabled: boolean, password: string): Promise<Guest>;
  deleteGuest(id: string): Promise<void>;
  members(): Promise<MemberList>;
  createMember(name: string, password: string): Promise<Member>;
  updateMember(id: string, disabled: boolean, password: string, role?: 'member' | 'admin'): Promise<Member>;
  deleteMember(id: string): Promise<void>;
  groups(): Promise<Group[]>;
  createGroup(name: string, accounts: string[]): Promise<Group>;
  updateGroup(id: string, name: string, accounts: string[]): Promise<Group>;
  deleteGroup(id: string): Promise<void>;
  privateByDefault(): Promise<boolean>;
  setPrivateByDefault(on: boolean): Promise<void>;
  kiosks(): Promise<KioskDisplay[]>;
  createKiosk(input: KioskInput): Promise<{ kiosk: KioskDisplay; token: string }>;
  updateKiosk(id: string, input: KioskInput): Promise<KioskDisplay>;
  renewKioskLink(id: string): Promise<string>;
  deleteKiosk(id: string): Promise<void>;
  kioskLogin(token: string): Promise<Session>;
  kioskOptions(): Promise<KioskOptions>;
  invites(): Promise<Invite[]>;
  createInvite(input: InviteInput): Promise<{ invite: Invite; token: string }>;
  revokeInvite(id: string): Promise<void>;
  createResetLink(accountId: string): Promise<{ token: string; expiresAt: Date }>;
  requestPasswordReset(name: string): Promise<void>;
  myEmail(): Promise<string>;
  setMyEmail(email: string): Promise<string>;
  inspectLink(token: string): Promise<LinkInfo>;
  acceptLink(token: string, password: string): Promise<LinkInfo>;
  apiTokens(): Promise<ApiToken[]>;
  createApiToken(name: string, scopes: string[], expiresInDays: number): Promise<string>;
  revokeApiToken(id: string): Promise<void>;
  /** Only accepted from a trusted reverse proxy that sets the user header. */
  headerLogin(): Promise<Session | { ticket: string }>;
  ssoSecondFactor(ticket: string, code: string): Promise<Session>;
  logout(): Promise<void>;
  changePassword(currentPassword: string, newPassword: string): Promise<Session>;
  sessions(): Promise<SessionInfo[]>;
  securityEvents(): Promise<SecurityEvent[]>;
  revokeSession(id: string): Promise<void>;
  revokeOtherSessions(): Promise<number>;
  newPairingCode(): Promise<PairingCode>;
  newHandoverCode(challenge: string): Promise<PairingCode>;
  twoFactorStatus(): Promise<TwoFactorStatus>;
  startTwoFactor(): Promise<TwoFactorEnrollment>;
  confirmTwoFactor(code: string): Promise<string[]>;
  disableTwoFactor(password: string, code: string): Promise<void>;
}

/** Output port: receives the CSRF token so later requests can send it. */
export interface CsrfTokenStore {
  setCsrfToken(token: string): void;
}

/** Output port: the browser's WebAuthn API (isolated so use cases stay testable). */
export interface PasskeyAuthenticator {
  supported(): boolean;
  create(options: PasskeyRegistrationOptions): Promise<PasskeyAttestation>;
  get(challenge: SecondFactorChallenge): Promise<PasskeyAssertion>;
}
