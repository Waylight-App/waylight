import type { ServerMessageKey } from '../../../api-client/generated/server-messages/wiki';

/**
 * German translations of the backend error messages of `wiki` (#102, #193).
 * The keys come from the generated catalogue (run `make generate` first); the
 * type makes a missing translation a compile error.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 5 wikis can be connected.
  srv_f0229a7c2d8c1f57: 'Es können höchstens 5 Wikis verbunden werden.',
  // Choose BookStack, Outline or Wiki.js.
  srv_a516606c590e59bd: 'Wähle BookStack, Outline oder Wiki.js.',
  // Enter the access token of the wiki.
  srv_19efb0d1db98099f: 'Gib das Zugangstoken des Wikis ein.',
  // The wiki is not connected (any more).
  srv_7d42213f5b873092: 'Das Wiki ist nicht (mehr) verbunden.',
};
