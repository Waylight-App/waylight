/**
 * A sign-in the Waylight app started in the browser (#560). The app opens
 * `/app-sign-in?challenge=…&state=…`; after the person signed in here and
 * confirmed, the page sends a single-use code back to the app.
 */
export interface AppHandover {
  /** S256 PKCE challenge of the app; only its verifier redeems the code. */
  challenge: string;
  /** Opaque value of the app, returned unchanged so it can match the answer. */
  state: string;
}

/** Path the app opens in the browser. */
export const appHandoverPath = '/app-sign-in';

const challengePattern = /^[A-Za-z0-9_-]{43}$/;
const statePattern = /^[A-Za-z0-9._~-]{1,128}$/;

/** Reads a handover from the query of the sign-in link; null if invalid. */
export function parseAppHandover(search: string): AppHandover | null {
  const params = new URLSearchParams(search);
  const challenge = params.get('challenge') ?? '';
  const state = params.get('state') ?? '';
  return challengePattern.test(challenge) && statePattern.test(state) ? { challenge, state } : null;
}

/**
 * The link back to the app. Only the app's own scheme is ever used, so the
 * page cannot be turned into a redirect to other sites.
 */
export function appCallback(
  handover: AppHandover,
  result: { code: string } | { error: 'access_denied' },
): string {
  const params = new URLSearchParams({ ...result, state: handover.state });
  return `waylight://auth?${params}`;
}
