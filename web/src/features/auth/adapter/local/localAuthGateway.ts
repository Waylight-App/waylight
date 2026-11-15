import type { AuthStatus } from '../../domain/auth';
import type { AuthGateway, CsrfTokenStore } from '../../usecase/ports';

/**
 * Raised by every account operation in the extension's local mode (ADR 0019).
 * There is no server and no account: the browser profile is the only user.
 */
export class LocalModeUnavailableError extends Error {
  override name = 'LocalModeUnavailableError';
  constructor() {
    super('This needs a trisse-home server.');
  }
}

const unavailable = (): Promise<never> => Promise.reject(new LocalModeUnavailableError());

/**
 * Auth gateway for the local mode of the browser extension. The person who
 * installed the extension is signed in as the owner from the start; sign-in,
 * guests, members, passkeys and every other account feature need a server
 * and reject with LocalModeUnavailableError.
 */
export class LocalAuthGateway implements AuthGateway {
  status(): Promise<AuthStatus> {
    return Promise.resolve({
      setupRequired: false,
      authenticated: true,
      // No server, so nothing to protect against cross-site requests.
      csrfToken: null,
      demo: false,
      guest: false,
      kiosk: false,
      admin: true,
      member: false,
      sso: { oidc: false, header: false, localLogin: false },
    });
  }

  // Signing out has nothing to end; the app reloads into the same state.
  logout(): Promise<void> {
    return Promise.resolve();
  }

  setup = unavailable;
  login = unavailable;
  passkeyLogin = unavailable;
  passkeys = unavailable;
  beginPasskeyRegistration = unavailable;
  finishPasskeyRegistration = unavailable;
  deletePasskey = unavailable;
  guests = unavailable;
  createGuest = unavailable;
  updateGuest = unavailable;
  deleteGuest = unavailable;
  members = unavailable;
  createMember = unavailable;
  updateMember = unavailable;
  deleteMember = unavailable;
  groups = unavailable;
  createGroup = unavailable;
  updateGroup = unavailable;
  deleteGroup = unavailable;
  privateByDefault = unavailable;
  setPrivateByDefault = unavailable;
  kiosks = unavailable;
  createKiosk = unavailable;
  updateKiosk = unavailable;
  renewKioskLink = unavailable;
  deleteKiosk = unavailable;
  kioskLogin = unavailable;
  kioskOptions = unavailable;
  apiTokens = unavailable;
  createApiToken = unavailable;
  revokeApiToken = unavailable;
  headerLogin = unavailable;
  ssoSecondFactor = unavailable;
  demoLogin = unavailable;
  changePassword = unavailable;
  sessions = unavailable;
  securityEvents = unavailable;
  revokeSession = unavailable;
  revokeOtherSessions = unavailable;
  twoFactorStatus = unavailable;
  startTwoFactor = unavailable;
  confirmTwoFactor = unavailable;
  disableTwoFactor = unavailable;
}

/** CSRF store for the local mode: there are no requests that need a token. */
export const noCsrf: CsrfTokenStore = {
  setCsrfToken() {
    // Intentionally empty, see LocalAuthGateway.status.
  },
};
