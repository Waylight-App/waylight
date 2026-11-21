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
};
