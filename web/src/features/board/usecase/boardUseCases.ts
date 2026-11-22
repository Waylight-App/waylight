import type { ShareGroup, Visibility } from '../../../shared/sharing/sharing';
import type { BoardNote, BoardNoteInput } from '../domain/board';

/** Output port: family board backend. */
export interface BoardGateway {
  list(): Promise<BoardNote[]>;
  create(input: BoardNoteInput, visibility?: Visibility): Promise<BoardNote>;
  update(id: string, input: BoardNoteInput, visibility?: Visibility): Promise<BoardNote>;
  remove(id: string): Promise<void>;
  shareGroups(): Promise<ShareGroup[]>;
}

const clean = (i: BoardNoteInput): BoardNoteInput => ({
  text: i.text.trim(),
  pinned: i.pinned,
  expiresOn: i.expiresOn,
});

/** Application logic of the family board (#685). */
export class BoardUseCases {
  constructor(private readonly gateway: BoardGateway) {}

  list(): Promise<BoardNote[]> {
    return this.gateway.list();
  }

  /** Writes (no id) or changes a note. */
  save(id: string | null, input: BoardNoteInput, visibility?: Visibility): Promise<BoardNote> {
    return id
      ? this.gateway.update(id, clean(input), visibility)
      : this.gateway.create(clean(input), visibility);
  }

  /** Pins or unpins one of the viewer's notes. */
  setPinned(note: BoardNote, pinned: boolean): Promise<BoardNote> {
    return this.gateway.update(note.id, { text: note.text, pinned, expiresOn: note.expiresOn });
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return this.gateway.shareGroups();
  }
}
