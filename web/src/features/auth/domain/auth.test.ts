import { describe, expect, it } from 'vitest';
import { redirectToProvider, remainingTime, validateNewPassword, viewFor } from './auth';

describe('viewFor', () => {
  it('prefers setup over everything else', () => {
    expect(
      viewFor({
        setupRequired: true,
        authenticated: false,
        csrfToken: null,
        demo: false,
        demoPassword: '',
        guest: false,
        kiosk: false,
        admin: false,
        owner: false,
        username: '',
        member: false,
        sso: { oidc: false, header: false, localLogin: true },
        passwordReset: false,
      }),
    ).toBe('setup');
  });
  it('shows the app for authenticated users', () => {
    expect(
      viewFor({
        setupRequired: false,
        authenticated: true,
        csrfToken: 'x',
        demo: false,
        demoPassword: '',
        guest: false,
        kiosk: false,
        admin: false,
        owner: false,
        username: '',
        member: false,
        sso: { oidc: false, header: false, localLogin: true },
        passwordReset: false,
      }),
    ).toBe('app');
  });
  it('shows the login otherwise', () => {
    expect(
      viewFor({
        setupRequired: false,
        authenticated: false,
        csrfToken: null,
        demo: false,
        demoPassword: '',
        guest: false,
        kiosk: false,
        admin: false,
        owner: false,
        username: '',
        member: false,
        sso: { oidc: false, header: false, localLogin: true },
        passwordReset: false,
      }),
    ).toBe('login');
  });
});

describe('validateNewPassword', () => {
  it.each([
    ['short', 'short', /at least/],
    ['            ', '            ', /whitespace/],
    ['long-enough-password', 'different-password', /match/],
  ])('rejects %j', (pw, confirm, want) => {
    expect(validateNewPassword(pw, confirm)).toMatch(want);
  });
  it('accepts a good password', () => {
    expect(validateNewPassword('long-enough-password', 'long-enough-password')).toBeNull();
  });
});

describe('kiosk displays (#50)', () => {
  it('puts the token into the fragment and knows night windows over midnight', async () => {
    const { inNightWindow, kioskLink } = await import('./auth');
    expect(kioskLink('https://home.example', 'a/b')).toBe('https://home.example/kiosk#a%2Fb');
    const at = (h: number, m = 0) => new Date(2026, 10, 14, h, m);
    expect(inNightWindow(at(23), '22:00', '06:30')).toBe(true);
    expect(inNightWindow(at(6, 15), '22:00', '06:30')).toBe(true);
    expect(inNightWindow(at(6, 30), '22:00', '06:30')).toBe(false);
    expect(inNightWindow(at(13), '12:00', '14:00')).toBe(true);
    expect(inNightWindow(at(13), '', '')).toBe(false);
  });
});

describe('remainingTime', () => {
  const now = new Date('2026-11-16T12:00:00Z');
  const at = (h: number) => new Date(now.getTime() + h * 3_600_000);

  it('counts whole days and hours on the last day', () => {
    expect(remainingTime(at(3 * 24 + 5), now)).toEqual({ unit: 'days', n: 3 });
    expect(remainingTime(at(24), now)).toEqual({ unit: 'days', n: 1 });
    expect(remainingTime(at(5.2), now)).toEqual({ unit: 'hours', n: 6 });
    expect(remainingTime(at(0.1), now)).toEqual({ unit: 'hours', n: 1 });
  });

  it('reports expired invitations', () => {
    expect(remainingTime(at(-1), now)).toEqual({ unit: 'expired' });
  });
});

describe('redirect to the provider (#514)', () => {
  const sso = { oidc: true, header: false, localLogin: false, autoRedirect: true };
  const q = (s = '') => new URLSearchParams(s);

  it('hands over only without a password form', () => {
    expect(redirectToProvider(sso, q())).toBe(true);
    expect(redirectToProvider({ ...sso, localLogin: true }, q())).toBe(false);
    expect(redirectToProvider({ ...sso, autoRedirect: false }, q())).toBe(false);
    expect(redirectToProvider({ ...sso, oidc: false }, q())).toBe(false);
  });

  it('never loops after the provider answered, and ?local=1 keeps the page', () => {
    expect(redirectToProvider(sso, q('sso_error=denied'))).toBe(false);
    expect(redirectToProvider(sso, q('sso_ticket=t'))).toBe(false);
    expect(redirectToProvider(sso, q('local=1'))).toBe(false);
  });
});
