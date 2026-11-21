import type { ServerMessageKey } from '../../../api-client/generated/server-messages/contracts';

/** German translations of the backend messages of `contracts` (#102, #687). */
export const serverDe: Record<ServerMessageKey, string> = {
  // Enter a cost between 0 and 1,000,000.
  srv_bbee068cd310e235: 'Bitte Kosten zwischen 0 und 1.000.000 eingeben.',
  // No deadlines ahead
  srv_c198ebaf810e6f0e: 'Keine Fristen in Sicht',
  // Remind between 0 and 90 days before the deadline.
  srv_a01290e0c0731c7d: 'Die Erinnerung kommt 0 bis 90 Tage vor der Frist.',
  // The contract does not exist.
  srv_9868ddfdf1b03925: 'Der Vertrag existiert nicht.',
  // The name must be between 1 and 80 characters long.
  srv_82e6a560a152e8bf: 'Der Name muss zwischen 1 und 80 Zeichen lang sein.',
  // The note may be at most 500 characters long.
  srv_279687ba82800dcf: 'Die Notiz darf höchstens 500 Zeichen lang sein.',
  // The notice period must be between 0 and 730 days.
  srv_a978dd17b95a872f: 'Die Kündigungsfrist muss zwischen 0 und 730 Tagen liegen.',
  // The period must be monthly, quarterly or yearly.
  srv_f06b513cc209bee6: 'Der Zeitraum muss monatlich, vierteljährlich oder jährlich sein.',
  // The provider may be at most 80 characters long.
  srv_08f974b98523102d: 'Der Anbieter darf höchstens 80 Zeichen lang sein.',
  // The renewal must be between 0 and 60 months.
  srv_a1b06e8dfdccdec0: 'Die Verlängerung muss zwischen 0 und 60 Monaten liegen.',
  // {name}: cancel by {date}
  srv_b9de50756b290a2b: '{name}: kündigen bis {date}',
};
