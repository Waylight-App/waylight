import type { ServerMessageKey } from '../../../api-client/generated/server-messages/countdown';

/**
 * German translations of the backend error messages of `countdown` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // graceDays must be a number.
  srv_1bef734e6f2cdc94: 'graceDays muss eine Zahl sein.',
  // The countdown does not exist.
  srv_1532da22068d27c4: 'Den Countdown gibt es nicht.',
  // The title must be between 1 and 120 characters long.
  srv_28a208881a6ef97d: 'Der Titel muss zwischen 1 und 120 Zeichen lang sein.',
  // How many days are left until a date, such as a holiday or a deadline.
  srv_e6e08c61b58d330d:
    'Wie viele Tage es noch bis zu einem Datum sind, etwa bis zum Urlaub oder einer Frist.',
  // Open the widget settings and add a date with a name.
  srv_2934b1f9d47467ca: 'Öffne die Widget-Einstellungen und füge ein Datum mit einem Namen hinzu.',
};
