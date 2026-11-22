/**
 * Paste from phone in the setup assistant: the assistant shows a QR code,
 * the signed-in mobile app answers it with the key copied on the phone, and
 * the server hands the key to the assistant once.
 */
import { getContext, setContext } from 'svelte';
import type { HttpClient } from '../api-client/http';
import type { SetupPasteDTO, SetupPasteStatusDTO } from '../api-client/generated/setupsignin-requests';
import { POLL_INTERVAL_MS, wait } from './signin';

/** A started code: what to poll and when it ends. */
export interface StartedPaste {
  code: string;
  expiresAt: Date;
}

/** The server side of a paste (output port of the assistant). */
export interface PasteGateway {
  start(): Promise<StartedPaste>;
  /** The key once the phone sent it, null while it has not. */
  take(code: string): Promise<string | null>;
  cancel(code: string): Promise<void>;
}

/** PasteGateway backed by /api/setup/paste. */
export class PasteApi implements PasteGateway {
  constructor(private readonly http: HttpClient) {}

  async start(): Promise<StartedPaste> {
    const p = await this.http.post<SetupPasteDTO>('/api/setup/paste', {});
    return { code: p.code, expiresAt: new Date(p.expiresAt) };
  }

  async take(code: string): Promise<string | null> {
    const s = await this.http.get<SetupPasteStatusDTO>(`/api/setup/paste/${encodeURIComponent(code)}`);
    return s.status === 'ready' && s.value ? s.value : null;
  }

  cancel(code: string): Promise<void> {
    return this.http.delete(`/api/setup/paste/${encodeURIComponent(code)}`);
  }
}

/** The text of the QR code; the app reads the address and the code from it. */
export function pasteLink(serverUrl: string, code: string): string {
  return `waylight://paste?${new URLSearchParams({ server: serverUrl, code })}`;
}

/** Raised when the code ended before the phone answered. */
export class PasteTimeout extends Error {
  constructor() {
    super('paste timed out');
    this.name = 'PasteTimeout';
  }
}

/**
 * Starts a code, hands it out and asks until the phone's key arrived. The
 * code ends on the server when it expires; an abort or an error cancels it
 * earlier.
 */
export async function followPaste(
  gateway: PasteGateway,
  show: (started: StartedPaste) => void,
  signal: AbortSignal,
  sleep: (ms: number, signal: AbortSignal) => Promise<void> = wait,
  now: () => number = Date.now,
): Promise<string> {
  const started = await gateway.start();
  if (signal.aborted) {
    void gateway.cancel(started.code).catch(() => {});
    throw signal.reason;
  }
  show(started);
  try {
    while (now() < started.expiresAt.getTime()) {
      await sleep(POLL_INTERVAL_MS, signal);
      const value = await gateway.take(started.code);
      if (value !== null) return value;
    }
    throw new PasteTimeout();
  } catch (err) {
    void gateway.cancel(started.code).catch(() => {});
    throw err;
  }
}

const key = Symbol('setup-paste');

/** Called once by the shell; without it the assistant offers no Paste from phone. */
export function providePaste(gateway: PasteGateway): void {
  setContext(key, gateway);
}

/** The paste gateway, if the shell provides one. */
export function usePaste(): PasteGateway | undefined {
  return getContext<PasteGateway | undefined>(key);
}
