/**
 * Signing in with the service from the setup assistant (#654) instead of
 * pasting a key: the server starts the sign-in, the user signs in in a new
 * window, the server receives the credential and keeps it. The browser only
 * sees the sign-in's ID and the page to open.
 */
import { getContext, setContext } from 'svelte';
import type { HttpClient } from '../api-client/http';
import type {
  SetupSignInDTO,
  SetupSignInInputDTO,
  SetupSignInStatusDTO,
} from '../api-client/generated/setupsignin-requests';
import type { SetupPreview } from './setup';

/** Services the server can sign in to. */
export type SignInService = 'nextcloud' | 'homeassistant';

/** How a descriptor offers signing in next to the pasted key. */
export interface SetupSignIn {
  service: SignInService;
  /** The product name on the buttons when it differs from the descriptor's, e.g. Nextcloud for CalDAV. */
  label?: string;
  /**
   * The service address the sign-in starts from; by default the address
   * field. Null hides the sign-in, e.g. when it cannot be derived.
   */
  address?: (values: Record<string, string>) => string | null;
  /** Saves the connection with the finished sign-in, like check. */
  check: (values: Record<string, string>, signIn: string) => Promise<SetupPreview>;
}

/** A started sign-in: what to poll and the page the user signs in on. */
export interface StartedSignIn {
  id: string;
  loginUrl: string;
}

/** The server side of a sign-in (output port of the assistant). */
export interface SignInGateway {
  start(service: SignInService, url: string): Promise<StartedSignIn>;
  /** Resolves "pending" or "ready"; a failed sign-in rejects with its reason. */
  status(id: string): Promise<'pending' | 'ready'>;
  cancel(id: string): Promise<void>;
}

/** SignInGateway backed by /api/setup/signin. */
export class SignInApi implements SignInGateway {
  constructor(private readonly http: HttpClient) {}

  start(service: SignInService, url: string): Promise<StartedSignIn> {
    const body: SetupSignInInputDTO = { service, url };
    return this.http.post<SetupSignInDTO>('/api/setup/signin', body);
  }

  async status(id: string): Promise<'pending' | 'ready'> {
    const s = await this.http.get<SetupSignInStatusDTO>(`/api/setup/signin/${encodeURIComponent(id)}`);
    return s.status === 'ready' ? 'ready' : 'pending';
  }

  cancel(id: string): Promise<void> {
    return this.http.delete(`/api/setup/signin/${encodeURIComponent(id)}`);
  }
}

/** How often the assistant asks the server; the server polls Nextcloud at most this often. */
export const POLL_INTERVAL_MS = 2000;

/**
 * Upper bound of polls: a little longer than the server keeps a sign-in
 * (ten minutes), so the server's "expired" answer normally ends it first.
 */
export const MAX_POLLS = 320;

/** Waits ms or rejects as soon as signal aborts. */
export function wait(ms: number, signal: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    if (signal.aborted) return reject(signal.reason);
    const timer = setTimeout(() => {
      signal.removeEventListener('abort', stop);
      resolve();
    }, ms);
    const stop = () => {
      clearTimeout(timer);
      reject(signal.reason);
    };
    signal.addEventListener('abort', stop, { once: true });
  });
}

/** Raised when a sign-in ran out of polls without an answer. */
export class SignInTimeout extends Error {
  constructor() {
    super('sign-in timed out');
    this.name = 'SignInTimeout';
  }
}

/**
 * Follows one sign-in: starts it, hands the page to open, then asks until
 * the server has the credential and resolves with the sign-in's ID. A
 * failure, a timeout or an abort (the assistant closed, the user chose the
 * key instead) cancels the sign-in on the server and rejects.
 */
export async function followSignIn(
  gateway: SignInGateway,
  service: SignInService,
  url: string,
  open: (loginUrl: string) => void,
  signal: AbortSignal,
  sleep: (ms: number, signal: AbortSignal) => Promise<void> = wait,
): Promise<string> {
  const started = await gateway.start(service, url);
  if (signal.aborted) {
    void gateway.cancel(started.id).catch(() => {});
    throw signal.reason;
  }
  open(started.loginUrl);
  try {
    for (let i = 0; i < MAX_POLLS; i++) {
      await sleep(POLL_INTERVAL_MS, signal);
      if ((await gateway.status(started.id)) === 'ready') return started.id;
    }
    throw new SignInTimeout();
  } catch (err) {
    void gateway.cancel(started.id).catch(() => {});
    throw err;
  }
}

const key = Symbol('setup-signin');

/** Called once by the shell; without it the assistant offers pasted keys only. */
export function provideSignIn(gateway: SignInGateway): void {
  setContext(key, gateway);
}

/** The sign-in gateway, if the shell provides one. */
export function useSignIn(): SignInGateway | undefined {
  return getContext<SignInGateway | undefined>(key);
}
