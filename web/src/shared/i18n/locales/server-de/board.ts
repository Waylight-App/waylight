import type { ServerMessageKey } from '../../../api-client/generated/server-messages/board';

/** German translations of the backend messages of `board` (#102, #685). */
export const serverDe: Record<ServerMessageKey, string> = {
  // 1 note
  srv_2af3ab7187e51be0: '1 Notiz',
  // A note must be between 1 and 500 characters long.
  srv_6446fa577a473ff5: 'Eine Notiz muss zwischen 1 und 500 Zeichen lang sein.',
  // Choose today or a later day.
  srv_46125f90eda403ad: 'Bitte heute oder einen späteren Tag wählen.',
  // The board is empty
  srv_69d75fbb47619635: 'Die Pinnwand ist leer',
  // {count} notes
  srv_8d7c330447920af1: '{count} Notizen',
  // {text}
  srv_819da5ee53f51a44: '{text}',
  // Short notes for the household, pinned or with a last day.
  srv_5b593a8f5989e9b3: 'Kurze Notizen für den Haushalt, angeheftet oder mit einem letzten Tag.',
  // Write a note with New note. Pin the ones that should stay on top and give short-lived ones a last day.
  srv_831ec27836aa7cd6:
    'Schreibe mit „Neue Notiz“ eine Notiz. Hefte an, was oben bleiben soll, und gib kurzlebigen einen letzten Tag.',
};
