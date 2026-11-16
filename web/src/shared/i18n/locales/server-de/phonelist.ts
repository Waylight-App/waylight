import type { ServerMessageKey } from '../../../api-client/generated/server-messages/phonelist';

/**
 * German translations of the backend error messages of `phonelist` (#102, #193).
 * The keys come from the generated catalogue (run `make generate` first); the
 * type makes a missing translation a compile error.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Enter a phone number or an e-mail address.
  srv_76ebaaaec60f0726: 'Gib eine Telefonnummer oder eine E-Mail-Adresse ein.',
  // Please enter a phone number with digits, spaces and + ( ) / . - only.
  srv_b7746e5bbafdfc41: 'Bitte gib eine Telefonnummer nur mit Ziffern, Leerzeichen und + ( ) / . - ein.',
  // The name must be between 1 and 100 characters long.
  srv_9f747d541533b228: 'Der Name muss zwischen 1 und 100 Zeichen lang sein.',
  // The phone list entry does not exist.
  srv_3425a1431cc501cb: 'Den Eintrag in der Telefonliste gibt es nicht.',
  // The phone list holds at most 1000 entries.
  srv_7941287f892c92ab: 'Die Telefonliste fasst höchstens 1000 Einträge.',
  // The role must be at most 100 characters long.
  srv_02bc94c990121a7f: 'Die Rolle darf höchstens 100 Zeichen lang sein.',
  // The team must be at most 60 characters long.
  srv_06810649bd9e2525: 'Das Team darf höchstens 60 Zeichen lang sein.',
  // This entry comes from the address book; change it there.
  srv_a10984d27954a094: 'Dieser Eintrag stammt aus dem Adressbuch; ändere ihn dort.',
};
