import type { ServerMessageKey } from '../../../api-client/generated/server-messages/places';

/**
 * German translations of the backend error messages of `places` (#102, #193).
 * The keys come from the generated catalogue (run `make generate` first); the
 * type makes a missing translation a compile error.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most {max} places can be saved.
  srv_605ec4bd7d27101f: 'Höchstens {max} Orte können gespeichert werden.',
  // The address of a place must be at most 200 characters long.
  srv_fc85c722d114f88b: 'Die Adresse eines Orts darf höchstens 200 Zeichen lang sein.',
  // The coordinates of the place are not valid.
  srv_6e715863e83f5e1d: 'Die Koordinaten des Orts sind ungültig.',
  // The name of a place must be between 1 and 80 characters long.
  srv_097e7261565de9b0: 'Der Name eines Orts muss zwischen 1 und 80 Zeichen lang sein.',
  // The place does not exist.
  srv_be110b760c6af4cc: 'Der Ort existiert nicht.',
};
