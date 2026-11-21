import type { ServerMessageKey } from '../../../api-client/generated/server-messages/habits';

/**
 * German translations of the backend error messages of `habits` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Only today and yesterday can be checked off.
  srv_a1fc697b9fe6467c: 'Abhaken geht nur für heute und gestern.',
  // The habit does not exist.
  srv_f37ac2a179a60cc1: 'Diese Gewohnheit gibt es nicht.',
  // No habits yet
  srv_3ae46504f4002e93: 'Noch keine Gewohnheiten',
  // {done} of {total} today
  srv_47c68f21f5a3748b: '{done} von {total} heute',
  // Done: {name}
  srv_1422d60397354b62: 'Erledigt: {name}',
  // Add habits in the widget and choose the days they are due; tap a habit to tick it off.
  srv_cab91ffc01eefcb8:
    'Füge Gewohnheiten im Widget hinzu und wähle die Tage, an denen sie fällig sind; tippe auf eine Gewohnheit, um sie abzuhaken.',
  // Your daily habits with streaks: tick off what you did today.
  srv_5b62b042bb03fd46: 'Deine täglichen Gewohnheiten mit Serien: hak ab, was du heute erledigt hast.',
};
