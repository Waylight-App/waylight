import { newId } from '../../../../shared/id/uuid';
import type { DocumentStore } from '../../../../shared/local/documentStore';
import type { ShareGroup } from '../../../../shared/sharing/sharing';
import {
  findByTitle,
  type DiffLine,
  type Note,
  type NoteInput,
  type NoteRef,
  type Revision,
  type RevisionDetail,
  type SearchHit,
  type Segment,
} from '../../domain/note';
import type { NotesGateway } from '../../usecase/notesUseCases';

/** A note as stored, in the shape of the server's notes table. */
interface NoteRow {
  id: string;
  title: string;
  body: string;
  pinned: boolean;
  createdAt: string;
  updatedAt: string;
  version: number;
}

/** An earlier content of a note, as the server's note_revisions table. */
interface RevisionRow {
  id: string;
  noteId: string;
  title: string;
  body: string;
  createdAt: string;
}

/** The notes document of the local mode. */
export interface NotesDoc {
  notes: NoteRow[];
  revisions: RevisionRow[];
}

// The server's limits (internal/modules/notes/domain/links.go): edits within
// ten minutes form one revision, and fifty are kept per note.
const REVISION_GAP_MS = 10 * 60 * 1000;
const MAX_REVISIONS = 50;
// Longer texts are shown as a full replacement, as on the server.
const MAX_DIFF_LINES = 3000;

const WIKI = /\[\[([^[\]|\n]+)(?:\|[^[\]\n]*)?\]\]/g;

class NotFoundError extends Error {
  override name = 'NotFoundError';
}

const toNote = (r: NoteRow): Note => ({
  id: r.id,
  title: r.title,
  body: r.body,
  pinned: r.pinned,
  updatedAt: new Date(r.updatedAt),
  version: r.version,
  visibility: 'private',
  mine: true,
});

const toRevision = (r: RevisionRow): Revision => ({
  id: r.id,
  title: r.title,
  createdAt: new Date(r.createdAt),
  size: r.body.length,
});

/** Line diff from a to b (longest common subsequence), like the server's DiffLines. */
export function diffLines(a: string, b: string): DiffLine[] {
  const x = a.split('\n');
  const y = b.split('\n');
  if (x.length > MAX_DIFF_LINES || y.length > MAX_DIFF_LINES) {
    return [
      ...x.map((text) => ({ op: '-' as const, text })),
      ...y.map((text) => ({ op: '+' as const, text })),
    ];
  }
  // lcs[i * w + j] = length of the LCS of x[i:] and y[j:]; a flat array
  // avoids nested index checks.
  const w = y.length + 1;
  const lcs = new Int32Array((x.length + 1) * w);
  const at = (i: number, j: number) => lcs[i * w + j] ?? 0;
  for (let i = x.length - 1; i >= 0; i--) {
    for (let j = y.length - 1; j >= 0; j--) {
      lcs[i * w + j] = x[i] === y[j] ? at(i + 1, j + 1) + 1 : Math.max(at(i + 1, j), at(i, j + 1));
    }
  }
  const out: DiffLine[] = [];
  let i = 0;
  let j = 0;
  while (i < x.length && j < y.length) {
    const xi = x[i] ?? '';
    const yj = y[j] ?? '';
    if (xi === yj) {
      out.push({ op: ' ', text: xi });
      i++;
      j++;
    } else if (at(i + 1, j) >= at(i, j + 1)) {
      out.push({ op: '-', text: xi });
      i++;
    } else {
      out.push({ op: '+', text: yj });
      j++;
    }
  }
  for (; i < x.length; i++) out.push({ op: '-', text: x[i] ?? '' });
  for (; j < y.length; j++) out.push({ op: '+', text: y[j] ?? '' });
  return out;
}

/** Splits text into segments with every case-insensitive match of the words marked. */
function highlight(text: string, words: string[]): Segment[] {
  if (words.length === 0) return [{ text, hit: false }];
  const escaped = words.map((w) => w.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
  const re = new RegExp(`(${escaped.join('|')})`, 'gi');
  return text
    .split(re)
    .filter((part) => part !== '')
    .map((part) => ({ text: part, hit: words.some((w) => w.toLowerCase() === part.toLowerCase()) }));
}

/** A short piece of the body around the first match. */
function snippetOf(body: string, words: string[]): string {
  const flat = body.replace(/\s+/g, ' ').trim();
  const lower = flat.toLowerCase();
  const at = Math.min(...words.map((w) => lower.indexOf(w.toLowerCase())).filter((n) => n >= 0), flat.length);
  const start = Math.max(0, at - 40);
  const piece = flat.slice(start, start + 140);
  return (start > 0 ? '…' : '') + piece + (start + 140 < flat.length ? '…' : '');
}

/**
 * Notes gateway of the browser extension's local mode (ADR 0019): notes and
 * their revisions in one document of the DocumentStore. Search, backlinks
 * and the revision diff work like the server's, in the browser. Sharing has
 * no meaning with a single local account, so every note is private.
 */
export class LocalNotesGateway implements NotesGateway {
  constructor(
    private readonly store: DocumentStore,
    private readonly now: () => Date = () => new Date(),
  ) {}

  private async read(): Promise<NotesDoc> {
    return (await this.store.get<NotesDoc>('notes')) ?? { notes: [], revisions: [] };
  }

  private async change<T>(fn: (doc: NotesDoc) => T): Promise<T> {
    const doc = await this.read();
    const result = fn(doc);
    await this.store.put('notes', doc);
    return result;
  }

  private static find(doc: NotesDoc, id: string): NoteRow {
    const row = doc.notes.find((n) => n.id === id);
    if (!row) throw new NotFoundError('The note does not exist.');
    return row;
  }

  async list(pinnedOnly: boolean): Promise<Note[]> {
    const { notes } = await this.read();
    return notes
      .filter((n) => !pinnedOnly || n.pinned)
      .sort((a, b) => Number(b.pinned) - Number(a.pinned) || b.updatedAt.localeCompare(a.updatedAt))
      .map(toNote);
  }

  async get(id: string): Promise<Note> {
    return toNote(LocalNotesGateway.find(await this.read(), id));
  }

  save(input: NoteInput, id?: string, opts?: { clientId?: string; version?: number }): Promise<Note> {
    return this.change((doc) => {
      const now = this.now().toISOString();
      if (!id) {
        const row: NoteRow = {
          id: opts?.clientId || newId(),
          ...input,
          createdAt: now,
          updatedAt: now,
          version: 1,
        };
        doc.notes.push(row);
        return toNote(row);
      }
      const row = LocalNotesGateway.find(doc, id);
      if (row.title !== input.title || row.body !== input.body) this.recordRevision(doc, row);
      Object.assign(row, input, { updatedAt: now, version: row.version + 1 });
      return toNote(row);
    });
  }

  /** Keeps the previous content unless a revision was taken moments ago. */
  private recordRevision(doc: NotesDoc, row: NoteRow): void {
    const mine = doc.revisions
      .filter((r) => r.noteId === row.id)
      .sort((a, b) => b.createdAt.localeCompare(a.createdAt));
    const newest = mine[0];
    const now = this.now();
    if (newest && now.getTime() - new Date(newest.createdAt).getTime() < REVISION_GAP_MS) return;
    doc.revisions.push({
      id: newId(),
      noteId: row.id,
      title: row.title,
      body: row.body,
      createdAt: now.toISOString(),
    });
    const prune = new Set(mine.slice(MAX_REVISIONS - 1).map((r) => r.id));
    if (prune.size) doc.revisions = doc.revisions.filter((r) => !prune.has(r.id));
  }

  delete(id: string): Promise<void> {
    return this.change((doc) => {
      doc.notes = doc.notes.filter((n) => n.id !== id);
      doc.revisions = doc.revisions.filter((r) => r.noteId !== id);
    });
  }

  async search(query: string): Promise<SearchHit[]> {
    const words = query.trim().split(/\s+/).filter(Boolean);
    if (words.length === 0) return [];
    const { notes } = await this.read();
    return notes
      .filter((n) => {
        const text = `${n.title}\n${n.body}`.toLowerCase();
        return words.every((w) => text.includes(w.toLowerCase()));
      })
      .sort((a, b) => b.updatedAt.localeCompare(a.updatedAt))
      .map((n) => ({
        id: n.id,
        title: highlight(n.title, words),
        snippet: highlight(snippetOf(n.body, words), words),
      }));
  }

  async backlinks(id: string): Promise<NoteRef[]> {
    const { notes } = await this.read();
    const target = LocalNotesGateway.find({ notes, revisions: [] }, id);
    return notes
      .filter((n) => n.id !== id)
      .filter((n) => [...n.body.matchAll(WIKI)].some((m) => findByTitle([target], m[1] ?? '')))
      .map((n) => ({ id: n.id, title: n.title }));
  }

  async revisions(id: string): Promise<Revision[]> {
    const { revisions } = await this.read();
    return revisions
      .filter((r) => r.noteId === id)
      .sort((a, b) => b.createdAt.localeCompare(a.createdAt))
      .map(toRevision);
  }

  async revision(id: string, revisionId: string): Promise<RevisionDetail> {
    const doc = await this.read();
    const note = LocalNotesGateway.find(doc, id);
    const rev = doc.revisions.find((r) => r.id === revisionId && r.noteId === id);
    if (!rev) throw new NotFoundError('The revision does not exist.');
    return { ...toRevision(rev), body: rev.body, diff: diffLines(rev.body, note.body) };
  }

  restore(id: string, revisionId: string): Promise<Note> {
    return this.change((doc) => {
      const row = LocalNotesGateway.find(doc, id);
      const rev = doc.revisions.find((r) => r.id === revisionId && r.noteId === id);
      if (!rev) throw new NotFoundError('The revision does not exist.');
      // The current content becomes a revision itself, so a restore can be undone.
      doc.revisions.push({
        id: newId(),
        noteId: id,
        title: row.title,
        body: row.body,
        createdAt: this.now().toISOString(),
      });
      Object.assign(row, {
        title: rev.title,
        body: rev.body,
        updatedAt: this.now().toISOString(),
        version: row.version + 1,
      });
      return toNote(row);
    });
  }

  async share(id: string): Promise<Note> {
    // One local account: nobody to share with, the note stays private.
    return this.get(id);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return Promise.resolve([]);
  }
}
