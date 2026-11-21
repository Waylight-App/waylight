import type { ServerMessageKey } from '../../../api-client/generated/server-messages/mail';

/** German translations of the backend messages of `mail` (#283). */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 10 mail accounts can be added.
  srv_a6a19c77984be1a4: 'Es können höchstens 10 E-Mail-Konten hinzugefügt werden.',
  // No mail account is set up yet.
  srv_f2d32e363bfff9e9: 'Es ist noch kein E-Mail-Konto eingerichtet.',
  // The mail account does not exist (any more).
  srv_1cd466bb63b42a23: 'Das E-Mail-Konto existiert nicht (mehr).',
  // Add a mailbox in Settings under Mail with the server name and your login.
  srv_e7d2e0fa24ebe178:
    'Füge in den Einstellungen unter E-Mail ein Postfach mit Servername und deinen Zugangsdaten hinzu.',
  // How many unread e-mails you have and the newest ones. Waylight only reads, it never sends or deletes.
  srv_ef77a55fb4d8a210:
    'Wie viele ungelesene E-Mails du hast und die neuesten davon. Waylight liest nur, es sendet oder löscht nie.',
};
