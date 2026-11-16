import type { ServerMessageKey } from '../../../api-client/generated/server-messages/shopping';

/**
 * German translations of the backend error messages of `shopping` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose text or ics.
  srv_3d5b8bd9dba6eec1: 'Wähle Text oder ics.',
  // Only the person who created the list can rename, share or delete it.
  srv_67340dd03bbd3030: 'Nur wer die Liste angelegt hat, kann sie umbenennen, teilen oder löschen.',
  // The list has 500 items; clear the ticked ones first.
  srv_b670eb0b48710065: 'Die Liste hat 500 Einträge; entferne zuerst die abgehakten.',
  // The shopping list does not exist or is not shared with you.
  srv_b112f7b441bebf0d: 'Die Einkaufsliste gibt es nicht oder sie ist nicht mit dir geteilt.',
  // You already have 20 shopping lists; delete one first.
  srv_3af5b06077e90463: 'Du hast schon 20 Einkaufslisten; lösche zuerst eine.',
  // Items must be between 1 and 200 characters long.
  srv_593a1c118a3d9d67: 'Einträge müssen zwischen 1 und 200 Zeichen lang sein.',
  // The item does not exist (any more).
  srv_77686a482211f04a: 'Der Eintrag existiert nicht (mehr).',
  // The item was changed in another app at the same time. Please try again.
  srv_7faea27f96c82f5f:
    'Der Eintrag wurde gleichzeitig in einer anderen App geändert. Bitte erneut versuchen.',
  // The shopping list is not configured yet.
  srv_7d7ee048ea53fd65: 'Die Einkaufsliste ist noch nicht eingerichtet.',
};
