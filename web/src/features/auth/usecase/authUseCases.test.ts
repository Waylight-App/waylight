import { describe, expect, it, vi } from 'vitest';
import type { AuthStatus, Session } from '../domain/auth';
import { AuthUseCases, ValidationError } from './authUseCases';
import type { AuthGateway } from './ports';

const session: Session = { csrfToken: 'csrf-1', expiresAt: new Date() };

function setup(status: Partial<AuthStatus> = {}) {
  const gateway: AuthGateway = {
    status: vi.fn().mockResolvedValue({
      setupRequired: false,
      authenticated: true,
      csrfToken: 'csrf-0',
      demo: false,
      guest: false,
      kiosk: false,
      admin: false,
      owner: false,
      member: false,
      sso: { oidc: false, header: false, localLogin: true },
      passwordReset: false,
      ...status,
    }),
    setup: vi.fn().mockResolvedValue(session),
    login: vi.fn().mockResolvedValue(session),
    demoLogin: vi.fn().mockResolvedValue(session),
    headerLogin: vi.fn().mockResolvedValue(session),
    passkeyLogin: vi.fn().mockResolvedValue(session),
    passkeys: vi.fn().mockResolvedValue({ available: true, passkeys: [] }),
    beginPasskeyRegistration: vi.fn().mockResolvedValue({
      challenge: 'c',
      rpId: 'h',
      userId: 'u',
      userName: 'owner',
      excludeCredentials: [],
    }),
    finishPasskeyRegistration: vi.fn().mockResolvedValue(['abcd-efgh']),
    deletePasskey: vi.fn(),
    ssoSecondFactor: vi.fn().mockResolvedValue(session),
    guests: vi.fn().mockResolvedValue([]),
    createGuest: vi.fn(),
    updateGuest: vi.fn(),
    deleteGuest: vi.fn(),
    members: vi.fn(),
    createMember: vi.fn(),
    updateMember: vi.fn(),
    deleteMember: vi.fn(),
    groups: vi.fn(),
    createGroup: vi.fn(),
    updateGroup: vi.fn(),
    deleteGroup: vi.fn(),
    privateByDefault: vi.fn(),
    setPrivateByDefault: vi.fn(),
    kiosks: vi.fn(),
    createKiosk: vi.fn().mockResolvedValue({ kiosk: {}, token: 't' }),
    updateKiosk: vi.fn(),
    renewKioskLink: vi.fn(),
    deleteKiosk: vi.fn(),
    kioskLogin: vi.fn().mockResolvedValue({ csrfToken: 'c' }),
    kioskOptions: vi.fn(),
    invites: vi.fn().mockResolvedValue([]),
    createInvite: vi.fn().mockResolvedValue({ invite: {}, token: 't' }),
    revokeInvite: vi.fn(),
    createResetLink: vi.fn(),
    inspectLink: vi.fn(),
    requestPasswordReset: vi.fn(),
    myEmail: vi.fn().mockResolvedValue(''),
    setMyEmail: vi.fn(),
    acceptLink: vi.fn().mockResolvedValue({ purpose: 'invite', name: 'lena' }),
    apiTokens: vi.fn().mockResolvedValue([]),
    createApiToken: vi.fn().mockResolvedValue('th_secret'),
    revokeApiToken: vi.fn(),
    logout: vi.fn().mockResolvedValue(undefined),
    changePassword: vi.fn().mockResolvedValue(session),
    securityEvents: vi.fn().mockResolvedValue([]),
    sessions: vi.fn().mockResolvedValue([
      {
        id: 'a',
        userAgent: '',
        clientIp: '',
        createdAt: new Date(1),
        lastSeenAt: new Date(5),
        current: false,
      },
      {
        id: 'b',
        userAgent: '',
        clientIp: '',
        createdAt: new Date(1),
        lastSeenAt: new Date(1),
        current: true,
      },
    ]),
    revokeSession: vi.fn(),
    revokeOtherSessions: vi.fn(),
    twoFactorStatus: vi.fn(),
    startTwoFactor: vi.fn(),
    confirmTwoFactor: vi.fn().mockResolvedValue(['a']),
    disableTwoFactor: vi.fn(),
  };
  const csrf = { setCsrfToken: vi.fn() };
  return { uc: new AuthUseCases(gateway, csrf), gateway, csrf };
}

describe('AuthUseCases', () => {
  it('signs in automatically on a demo instance', async () => {
    const { uc, gateway, csrf } = setup({ authenticated: false, csrfToken: null, demo: true });
    expect(await uc.currentView()).toBe('app');
    expect(gateway.demoLogin).toHaveBeenCalled();
    expect(csrf.setCsrfToken).toHaveBeenCalledWith('csrf-1');
    expect(uc.demo).toBe(true);
  });

  it('validates new guests before calling the backend', async () => {
    const { uc, gateway } = setup();
    await expect(uc.createGuest('Not Valid', 'guest-password-1', 'guest-password-1', [])).rejects.toThrow();
    await expect(uc.createGuest('anna', 'short', 'short', [])).rejects.toThrow();
    expect(gateway.createGuest).not.toHaveBeenCalled();
    await uc.createGuest(' Anna ', 'guest-password-1', 'guest-password-1', ['t1']);
    expect(gateway.createGuest).toHaveBeenCalledWith('anna', 'guest-password-1', ['t1']);
  });

  it('signs in through a trusted proxy header automatically', async () => {
    const { uc, gateway } = setup({
      authenticated: false,
      csrfToken: null,
      sso: { oidc: false, header: true, localLogin: true },
      passwordReset: false,
    });
    expect(await uc.currentView()).toBe('app');
    expect(gateway.headerLogin).toHaveBeenCalled();
  });

  it('keeps the SSO ticket when the local code is required', async () => {
    const { uc, gateway } = setup({
      authenticated: false,
      csrfToken: null,
      sso: { oidc: false, header: true, localLogin: true },
      passwordReset: false,
    });
    vi.mocked(gateway.headerLogin).mockResolvedValue({ ticket: 't1' });
    expect(await uc.currentView()).toBe('login');
    expect(uc.ssoTicket).toBe('t1');
    await uc.completeSso('t1', '123456');
    expect(gateway.ssoSecondFactor).toHaveBeenCalledWith('t1', '123456');
  });

  it('runs the passkey step when the server asks for it', async () => {
    const { gateway, csrf } = setup();
    const challenge = { ticket: 't', challenge: 'c', rpId: 'h', allowCredentials: ['id'], totp: false };
    vi.mocked(gateway.login).mockResolvedValue(challenge);
    const authenticator = {
      supported: () => true,
      create: vi.fn().mockResolvedValue({ clientDataJSON: 'cd', attestationObject: 'ao' }),
      get: vi.fn().mockResolvedValue({
        credentialId: 'id',
        clientDataJSON: 'cd',
        authenticatorData: 'ad',
        signature: 's',
      }),
    };
    const uc = new AuthUseCases(gateway, csrf, authenticator);
    expect(await uc.login('pw')).toEqual(challenge);
    await uc.loginWithPasskey(challenge);
    expect(gateway.passkeyLogin).toHaveBeenCalledWith('t', expect.objectContaining({ signature: 's' }));
    expect(csrf.setCsrfToken).toHaveBeenLastCalledWith('csrf-1');
    expect(await uc.addPasskey('Key')).toEqual(['abcd-efgh']);
    expect(authenticator.create).toHaveBeenCalled();
  });

  it('validates invitations and the password behind a link (#296)', async () => {
    const { uc, gateway } = setup();
    await expect(uc.createInvite({ name: 'A', role: 'guest', tabs: [], validDays: 7 })).rejects.toThrow(
      ValidationError,
    );
    await uc.createInvite({ name: ' Lena ', role: 'member', tabs: ['home'], validDays: 7 });
    expect(gateway.createInvite).toHaveBeenCalledWith({
      name: 'lena',
      role: 'member',
      tabs: [],
      validDays: 7,
    });

    await expect(uc.acceptLink('t', 'member-password-1', 'other-password-1')).rejects.toThrow(
      ValidationError,
    );
    expect(gateway.acceptLink).not.toHaveBeenCalled();
    await uc.acceptLink('t', 'member-password-1', 'member-password-1');
    expect(gateway.acceptLink).toHaveBeenCalledWith('t', 'member-password-1');
  });

  it('validates new members and group names before calling the backend', async () => {
    const { uc, gateway } = setup();
    await expect(uc.createMember('owner', 'member-password-1', 'member-password-1')).rejects.toThrow(
      ValidationError,
    );
    await expect(uc.createMember('ben', 'member-password-1', 'other-password-1')).rejects.toThrow(
      ValidationError,
    );
    expect(gateway.createMember).not.toHaveBeenCalled();
    await uc.createMember(' Ben ', 'member-password-1', 'member-password-1');
    expect(gateway.createMember).toHaveBeenCalledWith('ben', 'member-password-1');

    await expect(uc.createGroup('   ', [])).rejects.toThrow(ValidationError);
    await expect(uc.createGroup('x'.repeat(41), [])).rejects.toThrow(ValidationError);
    await uc.createGroup(' Office ', ['m1']);
    expect(gateway.createGroup).toHaveBeenCalledWith('Office', ['m1']);
  });

  it('remembers the admin and member roles', async () => {
    const member = setup({ member: true });
    await member.uc.currentView();
    expect(member.uc.member).toBe(true);
    expect(member.uc.admin).toBe(false);
    const admin = setup({ admin: true });
    await admin.uc.currentView();
    expect(admin.uc.admin).toBe(true);
  });

  it('remembers a guest session', async () => {
    const { uc } = setup({ guest: true });
    await uc.currentView();
    expect(uc.guest).toBe(true);
  });

  it('stores the csrf token from the status', async () => {
    const { uc, csrf } = setup();
    expect(await uc.currentView()).toBe('app');
    expect(csrf.setCsrfToken).toHaveBeenCalledWith('csrf-0');
  });

  it('validates the password before calling the backend', async () => {
    const { uc, gateway } = setup();
    await expect(uc.setup('token', 'short', 'short')).rejects.toBeInstanceOf(ValidationError);
    expect(gateway.setup).not.toHaveBeenCalled();
  });

  it('trims the setup token and stores the new csrf token', async () => {
    const { uc, gateway, csrf } = setup();
    await uc.setup('  token  ', 'long-enough-password', 'long-enough-password');
    expect(gateway.setup).toHaveBeenCalledWith('token', 'long-enough-password', 'everyday'); // safe default (#281)
    expect(csrf.setCsrfToken).toHaveBeenCalledWith('csrf-1');
    await uc.setup('token', 'long-enough-password', 'long-enough-password', 'homelab');
    expect(gateway.setup).toHaveBeenLastCalledWith('token', 'long-enough-password', 'homelab');
  });

  it('rejects an empty login password', async () => {
    const { uc } = setup();
    await expect(uc.login('')).rejects.toThrow(/password/);
  });

  it('clears the csrf token on logout', async () => {
    const { uc, csrf } = setup();
    await uc.logout();
    expect(csrf.setCsrfToken).toHaveBeenLastCalledWith('');
  });

  it('lists the current session first', async () => {
    const { uc } = setup();
    expect((await uc.sessions()).map((s) => s.id)).toEqual(['b', 'a']);
  });

  it('validates the confirmation code locally', async () => {
    const { uc, gateway } = setup();
    await expect(uc.confirmTwoFactor('12')).rejects.toBeInstanceOf(ValidationError);
    await uc.confirmTwoFactor('123 456');
    expect(gateway.confirmTwoFactor).toHaveBeenCalledWith('123456');
  });

  it('sends the second factor only when given', async () => {
    const { uc, gateway } = setup();
    await uc.login('pw', ' ');
    expect(gateway.login).toHaveBeenLastCalledWith('pw', undefined, undefined);
    await uc.login('pw', '123456');
    expect(gateway.login).toHaveBeenLastCalledWith('pw', '123456', undefined);
    await uc.login('pw', undefined, ' Anna ');
    expect(gateway.login).toHaveBeenLastCalledWith('pw', undefined, 'anna');
  });
});
