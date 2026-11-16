import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  EnrollmentDTO,
  RecoveryCodesDTO,
  RevokedDTO,
  SessionDTO,
  SessionInfoDTO,
  SSOTicketDTO,
  StatusDTO,
  TokenDTO,
  TwoFactorStatusDTO,
  CreatedTokenDTO,
  GuestDTO,
  GroupDTO,
  MemberDTO,
  PreferencesDTO,
  MembersDTO,
  EmailDTO,
  InviteCreatedDTO,
  InviteDTO,
  KioskCreatedDTO,
  KioskDTO,
  LinkInfoDTO,
  ResetLinkDTO,
  KioskLinkDTO,
  KioskOptionsDTO,
  PasskeysDTO,
  RegisteredPasskeyDTO,
  RegistrationOptionsDTO,
  SecondFactorDTO,
  SecurityEventDTO,
} from '../../../../shared/api-client/generated/auth';
import type {
  ChangePasswordRequestDTO,
  DisableTwoFactorDTO,
  GuestRequestDTO,
  GroupRequestDTO,
  MemberRequestDTO,
  PreferencesRequestDTO,
  LoginRequestDTO,
  PasskeyLoginDTO,
  PasskeyRegistrationDTO,
  SetupRequestDTO,
  SSOSecondFactorDTO,
  TokenRequestDTO,
  TwoFactorCodeDTO,
  AcceptLinkDTO,
  InviteRequestDTO,
  KioskLoginDTO,
  KioskRequestDTO,
  LinkTokenDTO,
  PasswordResetRequestDTO,
} from '../../../../shared/api-client/generated/auth-requests';
import type { ApiToken } from '../../domain/apiToken';
import type {
  Passkey,
  PasskeyAssertion,
  PasskeyAttestation,
  PasskeyRegistrationOptions,
  SecondFactorChallenge,
} from '../../domain/passkey';
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
  Member,
  MemberList,
  Session,
} from '../../domain/auth';
import type { TwoFactorEnrollment, TwoFactorStatus } from '../../domain/twoFactor';
import type { SecurityEvent } from '../../domain/securityEvent';
import type { SessionInfo } from '../../domain/session';
import type { AuthGateway } from '../../usecase/ports';

const toSession = (dto: SessionDTO): Session => ({
  csrfToken: dto.csrfToken,
  expiresAt: new Date(dto.expiresAt),
});

const toInvite = (d: InviteDTO): Invite => ({
  id: d.id,
  name: d.name,
  role: d.role === 'member' ? 'member' : 'guest',
  tabs: d.tabs ?? [],
  createdAt: new Date(d.createdAt),
  expiresAt: new Date(d.expiresAt),
});

const toLinkInfo = (d: LinkInfoDTO): LinkInfo => ({
  purpose: d.purpose === 'reset' ? 'reset' : 'invite',
  name: d.name,
});

const toKiosk = (d: KioskDTO): KioskDisplay => ({
  id: d.id,
  name: d.name,
  tabs: d.tabs,
  disabled: d.disabled,
  rotateSeconds: d.rotateSeconds,
  nightFrom: d.nightFrom,
  nightTo: d.nightTo,
  createdAt: new Date(d.createdAt),
  lastUsedAt: d.lastUsedAt ? new Date(d.lastUsedAt) : null,
});

const toMember = (d: MemberDTO): Member => ({
  id: d.id,
  name: d.name,
  groups: d.groups,
  disabled: d.disabled,
  createdAt: new Date(d.createdAt),
  role: d.role === 'admin' ? 'admin' : 'member',
  roleSource: d.roleSource ?? '',
});

const toGroup = (d: GroupDTO): Group => ({ id: d.id, name: d.name, accounts: d.accounts });

/** AuthGateway implementation backed by the REST API. */
export class AuthApi implements AuthGateway {
  constructor(private readonly http: HttpClient) {}

  async status(): Promise<AuthStatus> {
    const dto = await this.http.get<StatusDTO>('/api/auth/status');
    return {
      setupRequired: dto.setupRequired,
      authenticated: dto.authenticated,
      csrfToken: dto.csrfToken ?? null,
      demo: dto.demo,
      guest: dto.role === 'guest' || dto.role === 'kiosk',
      kiosk: dto.role === 'kiosk',
      admin: dto.role === 'owner' || dto.role === 'admin',
      owner: dto.role === 'owner',
      username: dto.username ?? '',
      member: dto.role === 'member',
      sso: dto.sso ?? { oidc: false, header: false, localLogin: true },
      passwordReset: dto.passwordReset ?? false,
    };
  }

  async demoLogin(): Promise<Session> {
    return toSession(await this.http.post<SessionDTO>('/api/auth/demo'));
  }

  async setup(setupToken: string, password: string, profile?: 'everyday' | 'homelab'): Promise<Session> {
    const body: SetupRequestDTO = { setupToken, password, ...(profile ? { profile } : {}) };
    return toSession(await this.http.post<SessionDTO>('/api/auth/setup', body));
  }

  async login(password: string, code?: string, username?: string): Promise<Session | SecondFactorChallenge> {
    const body: LoginRequestDTO = { password, ...(code ? { code } : {}), ...(username ? { username } : {}) };
    const res = await this.http.post<SessionDTO | SecondFactorDTO>('/api/auth/login', body);
    return 'secondFactor' in res
      ? { ...res.secondFactor, allowCredentials: res.secondFactor.allowCredentials ?? [] }
      : toSession(res);
  }

  async passkeyLogin(ticket: string, a: PasskeyAssertion): Promise<Session> {
    const body: PasskeyLoginDTO = { ticket, ...a };
    return toSession(await this.http.post<SessionDTO>('/api/auth/passkeys/login', body));
  }

  async passkeys(): Promise<{ available: boolean; passkeys: Passkey[] }> {
    const res = await this.http.get<PasskeysDTO>('/api/auth/passkeys');
    return {
      available: res.available,
      passkeys: (res.passkeys ?? []).map((p) => ({
        ...p,
        createdAt: new Date(p.createdAt),
        lastUsedAt: p.lastUsedAt ? new Date(p.lastUsedAt) : null,
      })),
    };
  }

  async beginPasskeyRegistration(): Promise<PasskeyRegistrationOptions> {
    const o = await this.http.post<RegistrationOptionsDTO>('/api/auth/passkeys/register/begin');
    return { ...o, excludeCredentials: o.excludeCredentials ?? [] };
  }

  async finishPasskeyRegistration(name: string, a: PasskeyAttestation): Promise<string[]> {
    const body: PasskeyRegistrationDTO = { name, ...a };
    return (
      (await this.http.post<RegisteredPasskeyDTO>('/api/auth/passkeys/register/finish', body))
        .recoveryCodes ?? []
    );
  }

  deletePasskey(id: string): Promise<void> {
    return this.http.delete(`/api/auth/passkeys/${encodeURIComponent(id)}`);
  }

  async logout(): Promise<void> {
    await this.http.post('/api/auth/logout');
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<Session> {
    const body: ChangePasswordRequestDTO = { currentPassword, newPassword };
    return toSession(await this.http.put<SessionDTO>('/api/auth/password', body));
  }

  async guests(): Promise<Guest[]> {
    return (await this.http.get<GuestDTO[]>('/api/auth/guests')).map(toGuest);
  }

  async createGuest(name: string, password: string, tabs: string[]): Promise<Guest> {
    const body: GuestRequestDTO = { name, password, tabs, disabled: false };
    return toGuest(await this.http.post<GuestDTO>('/api/auth/guests', body));
  }

  async updateGuest(id: string, tabs: string[], disabled: boolean, password: string): Promise<Guest> {
    const body: GuestRequestDTO = { tabs, disabled, ...(password ? { password } : {}) };
    return toGuest(await this.http.put<GuestDTO>(`/api/auth/guests/${encodeURIComponent(id)}`, body));
  }

  deleteGuest(id: string): Promise<void> {
    return this.http.delete(`/api/auth/guests/${encodeURIComponent(id)}`);
  }

  async members(): Promise<MemberList> {
    const r = await this.http.get<MembersDTO>('/api/auth/members');
    return { available: r.available, members: r.members.map(toMember) };
  }

  async createMember(name: string, password: string): Promise<Member> {
    const body: MemberRequestDTO = { name, password, disabled: false };
    return toMember(await this.http.post<MemberDTO>('/api/auth/members', body));
  }

  async updateMember(
    id: string,
    disabled: boolean,
    password: string,
    role?: 'member' | 'admin',
  ): Promise<Member> {
    const body: MemberRequestDTO = { disabled, ...(password ? { password } : {}), ...(role ? { role } : {}) };
    return toMember(await this.http.put<MemberDTO>(`/api/auth/members/${encodeURIComponent(id)}`, body));
  }

  deleteMember(id: string): Promise<void> {
    return this.http.delete(`/api/auth/members/${encodeURIComponent(id)}`);
  }

  async groups(): Promise<Group[]> {
    return (await this.http.get<GroupDTO[]>('/api/auth/groups')).map(toGroup);
  }

  async createGroup(name: string, accounts: string[]): Promise<Group> {
    const body: GroupRequestDTO = { name, accounts };
    return toGroup(await this.http.post<GroupDTO>('/api/auth/groups', body));
  }

  async updateGroup(id: string, name: string, accounts: string[]): Promise<Group> {
    const body: GroupRequestDTO = { name, accounts };
    return toGroup(await this.http.put<GroupDTO>(`/api/auth/groups/${encodeURIComponent(id)}`, body));
  }

  deleteGroup(id: string): Promise<void> {
    return this.http.delete(`/api/auth/groups/${encodeURIComponent(id)}`);
  }

  async privateByDefault(): Promise<boolean> {
    return (await this.http.get<PreferencesDTO>('/api/auth/preferences')).privateByDefault;
  }

  async setPrivateByDefault(on: boolean): Promise<void> {
    const body: PreferencesRequestDTO = { privateByDefault: on };
    await this.http.put<PreferencesDTO>('/api/auth/preferences', body);
  }

  async invites(): Promise<Invite[]> {
    return (await this.http.get<InviteDTO[]>('/api/auth/invites')).map(toInvite);
  }
  async createInvite(input: InviteInput): Promise<{ invite: Invite; token: string }> {
    const body: InviteRequestDTO = { ...input };
    const r = await this.http.post<InviteCreatedDTO>('/api/auth/invites', body);
    return { invite: toInvite(r.invite), token: r.token };
  }
  revokeInvite(id: string): Promise<void> {
    return this.http.delete(`/api/auth/invites/${encodeURIComponent(id)}`);
  }
  async createResetLink(accountId: string): Promise<{ token: string; expiresAt: Date }> {
    const r = await this.http.post<ResetLinkDTO>(
      `/api/auth/accounts/${encodeURIComponent(accountId)}/reset-link`,
      {},
    );
    return { token: r.token, expiresAt: new Date(r.expiresAt) };
  }
  requestPasswordReset(name: string): Promise<void> {
    const body: PasswordResetRequestDTO = { name };
    return this.http.post('/api/auth/password-reset', body);
  }
  async myEmail(): Promise<string> {
    return (await this.http.get<EmailDTO>('/api/auth/email')).email;
  }
  async setMyEmail(email: string): Promise<string> {
    const body: EmailDTO = { email };
    return (await this.http.put<EmailDTO>('/api/auth/email', body)).email;
  }
  async inspectLink(token: string): Promise<LinkInfo> {
    const body: LinkTokenDTO = { token };
    return toLinkInfo(await this.http.post<LinkInfoDTO>('/api/auth/link/inspect', body));
  }
  async acceptLink(token: string, password: string): Promise<LinkInfo> {
    const body: AcceptLinkDTO = { token, password };
    return toLinkInfo(await this.http.post<LinkInfoDTO>('/api/auth/link/accept', body));
  }
  async kiosks(): Promise<KioskDisplay[]> {
    return (await this.http.get<KioskDTO[]>('/api/auth/kiosks')).map(toKiosk);
  }
  async createKiosk(input: KioskInput): Promise<{ kiosk: KioskDisplay; token: string }> {
    const body: KioskRequestDTO = { ...input };
    const r = await this.http.post<KioskCreatedDTO>('/api/auth/kiosks', body);
    return { kiosk: toKiosk(r.kiosk), token: r.token };
  }
  async updateKiosk(id: string, input: KioskInput): Promise<KioskDisplay> {
    const body: KioskRequestDTO = { ...input };
    return toKiosk(await this.http.put<KioskDTO>(`/api/auth/kiosks/${encodeURIComponent(id)}`, body));
  }
  async renewKioskLink(id: string): Promise<string> {
    return (await this.http.post<KioskLinkDTO>(`/api/auth/kiosks/${encodeURIComponent(id)}/link`)).token;
  }
  deleteKiosk(id: string): Promise<void> {
    return this.http.delete(`/api/auth/kiosks/${encodeURIComponent(id)}`);
  }
  async kioskLogin(token: string): Promise<Session> {
    const body: KioskLoginDTO = { token };
    return toSession(await this.http.post<SessionDTO>('/api/auth/kiosk/login', body));
  }
  async kioskOptions(): Promise<KioskOptions> {
    return { ...(await this.http.get<KioskOptionsDTO>('/api/auth/kiosk')) };
  }

  async apiTokens(): Promise<ApiToken[]> {
    return (await this.http.get<TokenDTO[]>('/api/auth/tokens')).map(toApiToken);
  }

  async createApiToken(name: string, scopes: string[], expiresInDays: number): Promise<string> {
    const body: TokenRequestDTO = { name, scopes, expiresInDays };
    return (await this.http.post<CreatedTokenDTO>('/api/auth/tokens', body)).token;
  }

  revokeApiToken(id: string): Promise<void> {
    return this.http.delete(`/api/auth/tokens/${encodeURIComponent(id)}`);
  }

  async headerLogin(): Promise<Session | { ticket: string }> {
    const res = await this.http.post<SessionDTO | SSOTicketDTO>('/api/auth/sso/header');
    return 'ticket' in res ? { ticket: res.ticket } : toSession(res);
  }

  async ssoSecondFactor(ticket: string, code: string): Promise<Session> {
    const body: SSOSecondFactorDTO = { ticket, code };
    return toSession(await this.http.post<SessionDTO>('/api/auth/sso/2fa', body));
  }

  async securityEvents(): Promise<SecurityEvent[]> {
    const list = await this.http.get<SecurityEventDTO[]>('/api/auth/events');
    return list.map((e) => ({
      id: e.id,
      at: new Date(e.at),
      type: e.type,
      ip: e.ip,
      userAgent: e.userAgent,
      detail: e.detail,
    }));
  }

  async sessions(): Promise<SessionInfo[]> {
    const dtos = await this.http.get<SessionInfoDTO[]>('/api/auth/sessions');
    return dtos.map((d) => ({
      id: d.id,
      userAgent: d.userAgent,
      clientIp: d.clientIp,
      createdAt: new Date(d.createdAt),
      lastSeenAt: new Date(d.lastSeenAt),
      current: d.current,
    }));
  }

  revokeSession(id: string): Promise<void> {
    return this.http.delete(`/api/auth/sessions/${encodeURIComponent(id)}`);
  }

  async revokeOtherSessions(): Promise<number> {
    return (await this.http.post<RevokedDTO>('/api/auth/sessions/revoke-others')).revoked;
  }

  async twoFactorStatus(): Promise<TwoFactorStatus> {
    const d = await this.http.get<TwoFactorStatusDTO>('/api/auth/totp');
    return { enabled: d.enabled, recoveryCodesLeft: d.recoveryCodesLeft };
  }

  async startTwoFactor(): Promise<TwoFactorEnrollment> {
    const d = await this.http.post<EnrollmentDTO>('/api/auth/totp/setup');
    return { secret: d.secret, uri: d.uri };
  }

  async confirmTwoFactor(code: string): Promise<string[]> {
    const body: TwoFactorCodeDTO = { code };
    return (await this.http.post<RecoveryCodesDTO>('/api/auth/totp/confirm', body)).codes;
  }

  async disableTwoFactor(password: string, code: string): Promise<void> {
    const body: DisableTwoFactorDTO = { password, code };
    await this.http.post('/api/auth/totp/disable', body);
  }
}

const toGuest = (d: GuestDTO): Guest => ({ ...d, tabs: d.tabs ?? [], createdAt: new Date(d.createdAt) });

const toApiToken = (d: TokenDTO): ApiToken => ({
  ...d,
  scopes: d.scopes ?? [],
  createdAt: new Date(d.createdAt),
  expiresAt: d.expiresAt ? new Date(d.expiresAt) : null,
  lastUsedAt: d.lastUsedAt ? new Date(d.lastUsedAt) : null,
});
