import type { ServerMessageKey } from '../../../api-client/generated/server-messages/meters';

/** German translations of the backend messages of `meters` (#102, #689). */
export const serverDe: Record<ServerMessageKey, string> = {
  // Add a meter with its name and unit, then enter a reading with the day and the counter value. Enter one regularly, for example on the first of the month.
  srv_3a704529effbdd66:
    'Lege einen Zähler mit Name und Einheit an und trage dann eine Ablesung mit Tag und Zählerstand ein. Trage regelmäßig eine ein, zum Beispiel am Ersten des Monats.',
  // Choose electricity, gas, water, heat or other.
  srv_d0177e14188c1fd4: 'Wähle Strom, Gas, Wasser, Wärme oder Sonstiges.',
  // Enter a year like 2026.
  srv_bd7448b36f1e3d96: 'Gib ein Jahr wie 2026 ein.',
  // Enter the counter value, zero or more.
  srv_13433f458f53febb: 'Gib den Zählerstand ein, null oder mehr.',
  // No meters yet
  srv_e44d7b8aa0baf1f6: 'Noch keine Zähler',
  // No readings yet
  srv_b06ad407679a9b99: 'Noch keine Ablesungen',
  // The day cannot be in the future.
  srv_07a112349ed0ea95: 'Der Tag darf nicht in der Zukunft liegen.',
  // The meter does not exist.
  srv_702eba0dc8ce5082: 'Der Zähler existiert nicht.',
  // The reading cannot be higher than a later one.
  srv_1749452baf7b5543: 'Die Ablesung darf nicht höher sein als eine spätere.',
  // The reading cannot be lower than an earlier one.
  srv_8c2cceb976bc63a7: 'Die Ablesung darf nicht niedriger sein als eine frühere.',
  // The reading does not exist.
  srv_2981c9e71bd68e17: 'Die Ablesung existiert nicht.',
  // The unit must be between 1 and 12 characters long.
  srv_2252565573ce1db3: 'Die Einheit muss zwischen 1 und 12 Zeichen lang sein.',
  // Your electricity, gas, water and heat meters with the readings, the consumption per month and the consumption per year.
  srv_53d749df22ac762c:
    'Deine Strom-, Gas-, Wasser- und Wärmezähler mit den Ablesungen, dem Verbrauch pro Monat und dem Verbrauch pro Jahr.',
  // {name}: read 1 day ago
  srv_2ac876e33b16d18a: '{name}: vor 1 Tag abgelesen',
  // {name}: read today
  srv_e812d08f8989f733: '{name}: heute abgelesen',
  // {name}: read {days} days ago
  srv_a15f93c8ab35bf2f: '{name}: vor {days} Tagen abgelesen',
};
