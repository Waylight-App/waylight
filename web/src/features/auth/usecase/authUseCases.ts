import { t } from '../../../shared/i18n';
import type { ApiToken } from '../domain/apiToken';
import {
  type AuthView,
  type Group,
  type Guest,
  type Invite,
  type InviteInput,
  type LinkInfo,
  type Member,
  type MemberList,
  type KioskDisplay,
  type KioskInput,
  type KioskOptions,
  type SSOStatus,
  validGuestName,
  validateNewPassword,
  viewFor,
} from '../domain/auth';
import type { LinkedSignIn, MicrosoftFlow, MicrosoftPoll } from '../domain/identity';
import type { SecurityEvent } from '../domain/securityEvent';
import type { SessionInfo } from '../domain/session';
import type { PairingCode } from '../domain/appPairing';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../domain/twoFactor';
import type { Passkey, SecondFactorChallenge } from '../domain/passkey';
import type { AuthGateway, CsrfTokenStore, PasskeyAuthenticator } from './ports';

/** Error for input rejected before it reaches the backend. */
export class ValidationError extends Error {
  override name = 'ValidationError';
}

/** Application logic of the auth feature. */
export class AuthUseCases {
  constructor(
    private readonly gateway: AuthGateway,
    private readonly csrf: CsrfTokenStore,
    private readonly authenticator?: PasskeyAuthenticator,
  ) {}

  /** Whether this browser can use passkeys here (secure context). */
  passkeysSupported(): boolean {
    return this.authenticator?.supported() ?? false;
  }

  /** True on a public demo instance (set by currentView). */
  demo = false;
  /** Public password of the demo account, shown on the sign-in page. */
  demoPassword = '';
  /** True when a read-only guest is signed in (set by currentView). */
  guest = false;
  /** True when a kiosk display is signed in (#50); guest is true as well. */
  kiosk = false;
  /** True for an admin of the instance: the owner or an admin (ADR 0022). */
  admin = false;
  /** True for the owner, the one account that can never be removed (ADR 0022). */
  owner = false;
  /** Name of the signed-in account. */
  username = '';
  /** True for a member: own tabs and content, no instance settings (#294). */
  member = false;
  /** Sign-in methods offered by the server (set by currentView). */
  sso: SSOStatus = { oidc: false, header: false, localLogin: true };
  /** Reset links by e-mail are available (#296). */
  passwordReset = false;
  /** Ticket of an SSO sign-in that still needs the local code. */
  ssoTicket = '';

  /** Loads the status and returns the screen to show. */
  async currentView(): Promise<AuthView> {
    const status = await this.gateway.status();
    this.demo = status.demo;
    this.demoPassword = status.demoPassword;
    this.guest = status.guest;
    this.kiosk = status.kiosk;
    this.admin = status.admin;
    this.owner = status.owner;
    this.username = status.username;
    this.member = status.member;
    this.sso = status.sso;
    this.passwordReset = status.passwordReset;
    // Behind a reverse proxy with forward auth the user is already known.
    if (status.sso.header && !status.authenticated && !status.setupRequired) {
      try {
        const res = await this.gateway.headerLogin();
        if ('ticket' in res) this.ssoTicket = res.ticket;
        else {
          this.csrf.setCsrfToken(res.csrfToken);
          return 'app';
        }
      } catch {
        // Not allowed or not configured: fall back to the login page.
      }
    }
    if (status.csrfToken) this.csrf.setCsrfToken(status.csrfToken);
    return viewFor(status);
  }

  /** Creates the account with the setup token from the server log. */
  async setup(
    setupToken: string,
    password: string,
    confirmation: string,
    profile: 'everyday' | 'homelab' = 'everyday',
  ): Promise<void> {
    const problem = validateNewPassword(password, confirmation);
    if (problem) throw new ValidationError(problem);
    if (setupToken.trim() === '') throw new ValidationError(t('auth.enter_setup_token'));
    const session = await this.gateway.setup(setupToken.trim(), password, profile);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  /**
   * Logs in with the password and, if two-factor authentication is enabled,
   * a code. Guests also give their name.
   */
  async login(password: string, code?: string, username?: string): Promise<SecondFactorChallenge | null> {
    if (password === '') throw new ValidationError(t('auth.enter_password'));
    const name = username?.trim().toLowerCase() || undefined;
    const res = await this.gateway.login(password, code?.trim() || undefined, name);
    if ('ticket' in res) return res;
    this.csrf.setCsrfToken(res.csrfToken);
    return null;
  }

  /** Second login step: sign the challenge with a passkey. */
  async loginWithPasskey(challenge: SecondFactorChallenge): Promise<void> {
    if (!this.authenticator) throw new ValidationError(t('auth.passkeys_unsupported'));
    const assertion = await this.authenticator.get(challenge);
    const session = await this.gateway.passkeyLogin(challenge.ticket, assertion);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  passkeys(): Promise<{ available: boolean; passkeys: Passkey[] }> {
    return this.gateway.passkeys();
  }

  /** Registers a passkey; returns new recovery codes when they were issued. */
  async addPasskey(name: string): Promise<string[]> {
    if (!this.authenticator?.supported()) throw new ValidationError(t('auth.passkeys_unsupported'));
    const options = await this.gateway.beginPasskeyRegistration();
    const attestation = await this.authenticator.create(options);
    return this.gateway.finishPasskeyRegistration(name.trim(), attestation);
  }

  deletePasskey(id: string): Promise<void> {
    return this.gateway.deletePasskey(id);
  }

  /** Sign-ins at the OpenID provider linked to this account (#527). */
  linkedSignIns(): Promise<LinkedSignIn[]> {
    return this.gateway.linkedSignIns();
  }

  /** Starts linking a sign-in; returns the provider URL to open. */
  beginLinkSignIn(): Promise<string> {
    return this.gateway.beginLinkSignIn();
  }

  unlinkSignIn(subject: string): Promise<void> {
    return this.gateway.unlinkSignIn(subject);
  }

  /** Starts a Microsoft code flow: to link (signed in) or to sign in (#535). */
  startMicrosoft(link: boolean): Promise<MicrosoftFlow> {
    return this.gateway.startMicrosoft(link);
  }

  /** Polls a Microsoft code flow; a finished sign-in keeps its CSRF token. */
  async pollMicrosoft(flow: string): Promise<MicrosoftPoll> {
    const res = await this.gateway.pollMicrosoft(flow);
    if ('csrfToken' in res) {
      this.csrf.setCsrfToken(res.csrfToken);
      return { signedIn: true };
    }
    return res;
  }

  /** Accepts an invitation through the provider; returns the URL to open (#528). */
  signUpWithSSO(token: string): Promise<string> {
    return this.gateway.signUpWithSSO(token);
  }

  /** Finishes an SSO sign-in that needs the local authentication code. */
  async completeSso(ticket: string, code: string): Promise<void> {
    if (code.trim() === '') throw new ValidationError(t('auth.enter_code_hint'));
    const session = await this.gateway.ssoSecondFactor(ticket, code.trim());
    this.ssoTicket = '';
    this.csrf.setCsrfToken(session.csrfToken);
  }

  /** Ends the current session. */
  async logout(): Promise<void> {
    await this.gateway.logout();
    this.csrf.setCsrfToken('');
  }

  /** Changes the password; all other sessions are signed out by the backend. */
  async changePassword(current: string, next: string, confirmation: string): Promise<void> {
    const problem = validateNewPassword(next, confirmation);
    if (problem) throw new ValidationError(problem);
    const session = await this.gateway.changePassword(current, next);
    this.csrf.setCsrfToken(session.csrfToken);
  }

  /** Guest accounts (owner only). */
  guests(): Promise<Guest[]> {
    return this.gateway.guests();
  }

  /** Creates a read-only guest that sees the given tabs. */
  async createGuest(name: string, password: string, confirmation: string, tabs: string[]): Promise<Guest> {
    const n = name.trim().toLowerCase();
    if (!validGuestName(n)) throw new ValidationError(t('auth.guest_name_invalid'));
    const problem = validateNewPassword(password, confirmation);
    if (problem) throw new ValidationError(problem);
    return this.gateway.createGuest(n, password, tabs);
  }

  /** Changes shared tabs, the disabled flag and optionally the password. */
  async updateGuest(guest: Guest, tabs: string[], disabled: boolean, password = ''): Promise<Guest> {
    if (password !== '') {
      const problem = validateNewPassword(password, password);
      if (problem) throw new ValidationError(problem);
    }
    return this.gateway.updateGuest(guest.id, tabs, disabled, password);
  }

  deleteGuest(id: string): Promise<void> {
    return this.gateway.deleteGuest(id);
  }

  /** Member accounts (admin only, #294). */
  members(): Promise<MemberList> {
    return this.gateway.members();
  }

  /** Creates a member; names follow the guest rules and share their namespace. */
  async createMember(name: string, password: string, confirmation: string): Promise<Member> {
    const n = name.trim().toLowerCase();
    if (!validGuestName(n)) throw new ValidationError(t('auth.guest_name_invalid'));
    const problem = validateNewPassword(password, confirmation);
    if (problem) throw new ValidationError(problem);
    return this.gateway.createMember(n, password);
  }

  /** Enables, disables or sets a new password; an empty password keeps it. */
  async updateMember(member: Member, disabled: boolean, password = ''): Promise<Member> {
    if (password !== '') {
      const problem = validateNewPassword(password, password);
      if (problem) throw new ValidationError(problem);
    }
    return this.gateway.updateMember(member.id, disabled, password);
  }

  /** Makes a member an admin or back (ADR 0022); demotion signs them out. */
  setMemberRole(member: Member, role: 'member' | 'admin'): Promise<Member> {
    return this.gateway.updateMember(member.id, member.disabled, '', role);
  }

  /** Whether a member row is the signed-in account (admins cannot change their own role). */
  isSelf(member: Member): boolean {
    return member.name === this.username;
  }

  deleteMember(id: string): Promise<void> {
    return this.gateway.deleteMember(id);
  }

  /** Groups that records can be shared with (admin only). */
  groups(): Promise<Group[]> {
    return this.gateway.groups();
  }

  async createGroup(name: string, accounts: string[]): Promise<Group> {
    const n = name.trim();
    if (!n || n.length > 40) throw new ValidationError(t('auth.group_name_invalid'));
    return this.gateway.createGroup(n, accounts);
  }

  async updateGroup(group: Group, name: string, accounts: string[]): Promise<Group> {
    const n = name.trim();
    if (!n || n.length > 40) throw new ValidationError(t('auth.group_name_invalid'));
    return this.gateway.updateGroup(group.id, n, accounts);
  }

  deleteGroup(id: string): Promise<void> {
    return this.gateway.deleteGroup(id);
  }

  /** Whether the admin's own new entries start private (ADR 0018). */
  privateByDefault(): Promise<boolean> {
    return this.gateway.privateByDefault();
  }

  setPrivateByDefault(on: boolean): Promise<void> {
    return this.gateway.setPrivateByDefault(on);
  }

  /** Open invitations (admin only, #296). */
  invites(): Promise<Invite[]> {
    return this.gateway.invites();
  }

  /** Invites a member or a guest; the token is shown once. */
  async createInvite(input: InviteInput): Promise<{ invite: Invite; token: string }> {
    const name = input.name.trim().toLowerCase();
    if (!validGuestName(name)) throw new ValidationError(t('auth.guest_name_invalid'));
    return this.gateway.createInvite({ ...input, name, tabs: input.role === 'guest' ? input.tabs : [] });
  }

  revokeInvite(id: string): Promise<void> {
    return this.gateway.revokeInvite(id);
  }

  /** A link that lets a member or guest set a new password (24 hours). */
  createResetLink(accountId: string): Promise<{ token: string; expiresAt: Date }> {
    return this.gateway.createResetLink(accountId);
  }

  /** Asks for a reset link by e-mail; the answer is the same for every name. */
  async requestPasswordReset(name: string): Promise<void> {
    const n = name.trim().toLowerCase();
    if (!validGuestName(n)) throw new ValidationError(t('auth.guest_name_invalid'));
    return this.gateway.requestPasswordReset(n);
  }

  /** The own address for reset links (members and guests). */
  myEmail(): Promise<string> {
    return this.gateway.myEmail();
  }

  setMyEmail(email: string): Promise<string> {
    return this.gateway.setMyEmail(email.trim());
  }

  /** What an invitation or reset link is for; public. */
  inspectLink(token: string): Promise<LinkInfo> {
    return this.gateway.inspectLink(token);
  }

  /** Sets the password behind an invitation or reset link; public. */
  async acceptLink(token: string, password: string, confirmation: string): Promise<LinkInfo> {
    const problem = validateNewPassword(password, confirmation);
    if (problem) throw new ValidationError(problem);
    return this.gateway.acceptLink(token, password);
  }

  /** Kiosk displays (owner only, #50). */
  kiosks(): Promise<KioskDisplay[]> {
    return this.gateway.kiosks();
  }
  /** Creates a display; the link token is returned once. */
  async createKiosk(input: KioskInput): Promise<{ kiosk: KioskDisplay; token: string }> {
    const name = input.name.trim().toLowerCase();
    if (!validGuestName(name)) throw new ValidationError(t('auth.guest_name_invalid'));
    return this.gateway.createKiosk({ ...input, name });
  }
  updateKiosk(id: string, input: KioskInput): Promise<KioskDisplay> {
    return this.gateway.updateKiosk(id, input);
  }
  renewKioskLink(id: string): Promise<string> {
    return this.gateway.renewKioskLink(id);
  }
  deleteKiosk(id: string): Promise<void> {
    return this.gateway.deleteKiosk(id);
  }
  /** Signs this browser in as a kiosk display with the token of its link. */
  async kioskLogin(token: string): Promise<void> {
    const session = await this.gateway.kioskLogin(token);
    this.csrf.setCsrfToken(session.csrfToken);
  }
  /** Display options of the signed-in kiosk. */
  kioskOptions(): Promise<KioskOptions> {
    return this.gateway.kioskOptions();
  }

  /** Personal API tokens (owner only). */
  apiTokens(): Promise<ApiToken[]> {
    return this.gateway.apiTokens();
  }

  /** Creates a token and returns its secret, which is shown only once. */
  async createApiToken(name: string, scopes: string[], expiresInDays: number): Promise<string> {
    if (name.trim() === '') throw new ValidationError(t('auth.token_name_required'));
    if (scopes.length === 0) throw new ValidationError(t('auth.token_scope_required'));
    return this.gateway.createApiToken(name.trim(), scopes, Math.max(0, Math.round(expiresInDays) || 0));
  }

  revokeApiToken(id: string): Promise<void> {
    return this.gateway.revokeApiToken(id);
  }

  /** The security log, newest first. */
  securityEvents(): Promise<SecurityEvent[]> {
    return this.gateway.securityEvents();
  }

  /** Lists active sessions; the current one first. */
  async sessions(): Promise<SessionInfo[]> {
    const list = await this.gateway.sessions();
    return [...list].sort(
      (a, b) => Number(b.current) - Number(a.current) || b.lastSeenAt.getTime() - a.lastSeenAt.getTime(),
    );
  }

  revokeSession(id: string): Promise<void> {
    return this.gateway.revokeSession(id);
  }

  /** Signs out every other device; returns how many sessions ended. */
  revokeOtherSessions(): Promise<number> {
    return this.gateway.revokeOtherSessions();
  }

  /** A single-use code for the QR code that pairs the mobile app. */
  newPairingCode(): Promise<PairingCode> {
    return this.gateway.newPairingCode();
  }

  /** The code the sign-in page hands to the app after a browser sign-in. */
  newHandoverCode(challenge: string): Promise<PairingCode> {
    return this.gateway.newHandoverCode(challenge);
  }

  twoFactorStatus(): Promise<TwoFactorStatus> {
    return this.gateway.twoFactorStatus();
  }

  /** Starts connecting an authenticator app; 2FA is active only after confirmation. */
  startTwoFactor(): Promise<TwoFactorEnrollment> {
    return this.gateway.startTwoFactor();
  }

  /** Confirms the setup with a code and returns the one-time recovery codes. */
  confirmTwoFactor(code: string): Promise<string[]> {
    const c = code.replace(/\s/g, '');
    if (!/^\d{6}$/.test(c)) return Promise.reject(new ValidationError(t('auth.enter_code')));
    return this.gateway.confirmTwoFactor(c);
  }

  disableTwoFactor(password: string, code: string): Promise<void> {
    if (password === '' || code.trim() === '') {
      return Promise.reject(new ValidationError(t('auth.password_and_code_required')));
    }
    return this.gateway.disableTwoFactor(password, code.trim());
  }
}
