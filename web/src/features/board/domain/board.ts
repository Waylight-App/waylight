import type { Visibility } from '../../../shared/sharing/sharing';

/** A note on the family board (#685). */
export interface BoardNote {
  id: string;
  text: string;
  pinned: boolean;
  /** Last day the note is shown, YYYY-MM-DD, or empty. */
  expiresOn: string;
  visibility: Visibility;
  mine: boolean;
}

/** What the dialog edits. */
export interface BoardNoteInput {
  text: string;
  pinned: boolean;
  expiresOn: string;
}

/** Limit; it mirrors the backend. */
export const MAX_TEXT_LENGTH = 500;

/** Message keys of validateNote. */
export type NoteError = 'board.error.text' | 'board.error.expiry';

/** Validates the dialog against today (YYYY-MM-DD) and returns a message key or null. */
export function validateNote(input: BoardNoteInput, today: string): NoteError | null {
  const text = input.text.trim();
  if (!text || [...text].length > MAX_TEXT_LENGTH) return 'board.error.text';
  if (input.expiresOn && input.expiresOn < today) return 'board.error.expiry';
  return null;
}
