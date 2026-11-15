import type { ServerMessageKey } from '../../../api-client/generated/server-messages/mail';

/** German translations of the backend messages of `mail` (#283). */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 10 mail accounts can be added.
  srv_a6a19c77984be1a4: 'Es können höchstens 10 E-Mail-Konten hinzugefügt werden.',
  // No mail account is set up yet.
  srv_f2d32e363bfff9e9: 'Es ist noch kein E-Mail-Konto eingerichtet.',
  // The mail account does not exist (any more).
  srv_1cd466bb63b42a23: 'Das E-Mail-Konto existiert nicht (mehr).',
};
