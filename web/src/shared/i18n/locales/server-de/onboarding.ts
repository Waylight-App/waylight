import type { ServerMessageKey } from '../../../api-client/generated/server-messages/onboarding';

/**
 * German translations of the backend error messages of `onboarding` (#739). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The step is invalid.
  srv_ebf9b10ecfebaca5: 'Der Schritt ist ungültig.',
};
