import type { ServerMessageKey } from '../../../api-client/generated/server-messages/discovery';

/**
 * German translations of the backend error messages of `discovery` (#102, #193).
 * The keys come from the generated catalogue (run `make generate` first); the
 * type makes a missing translation a compile error.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // This integration cannot be found in the home network.
  srv_1f6f02017c5103f8: 'Diese Integration lässt sich im Heimnetz nicht finden.',
};
