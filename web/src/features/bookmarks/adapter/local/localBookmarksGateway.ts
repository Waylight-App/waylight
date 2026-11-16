import { newId } from '../../../../shared/id/uuid';
import type { DocumentStore } from '../../../../shared/local/documentStore';
import type { ShareGroup } from '../../../../shared/sharing/sharing';
import type { Bookmark, BookmarkInput, BookmarkSettings, Folder } from '../../domain/bookmark';
import type {
  BookmarkFilter,
  BookmarkImportResult,
  BookmarkList,
  BookmarksGateway,
} from '../../usecase/bookmarksUseCases';

/** A bookmark as stored, in the shape of the server's bookmarks table plus its tags. */
interface BookmarkRow {
  id: string;
  folderId: string;
  title: string;
  url: string;
  description: string;
  tags: string[];
  createdAt: string;
}

/** A folder as stored, in the shape of the server's bookmark_folders table. */
interface FolderRow {
  id: string;
  name: string;
  position: number;
}

/** The bookmarks document of the local mode. */
export interface BookmarksDoc {
  bookmarks: BookmarkRow[];
  folders: FolderRow[];
}

// The server's limits (internal/modules/bookmarks/domain/bookmark.go).
const MAX_TITLE = 200;
const MAX_DESCRIPTION = 1000;
const MAX_URL = 2048;
const MAX_TAGS = 10;
const MAX_TAG = 30;
const MAX_FOLDER_NAME = 60;
const TAG = /^[\p{L}\p{N}_-]+$/u;

/** Input the server would refuse; the message is the server's. */
export class BookmarkValidationError extends Error {
  override name = 'BookmarkValidationError';
}

class NotFoundError extends Error {
  override name = 'NotFoundError';
}

const chars = (s: string) => [...s].length;

/** Validates and normalises like the server's Bookmark.Update and NormaliseTags. */
function normalise(input: BookmarkInput): Omit<BookmarkRow, 'id' | 'createdAt'> {
  const url = input.url.trim();
  let parsed: URL | undefined;
  try {
    parsed = new URL(url);
  } catch {
    parsed = undefined;
  }
  if (
    !parsed ||
    (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') ||
    !parsed.host ||
    url.length > MAX_URL
  ) {
    throw new BookmarkValidationError('The URL must start with http:// or https://.');
  }
  const title = input.title.trim() || parsed.host;
  if (chars(title) > MAX_TITLE)
    throw new BookmarkValidationError('The title must be at most 200 characters long.');
  const description = input.description.trim();
  if (chars(description) > MAX_DESCRIPTION)
    throw new BookmarkValidationError('The description must be at most 1000 characters long.');
  const tags: string[] = [];
  for (const raw of input.tags) {
    const tag = raw.trim().replace(/^#/, '').toLowerCase();
    if (tag === '' || tags.includes(tag)) continue;
    if (chars(tag) > MAX_TAG || !TAG.test(tag))
      throw new BookmarkValidationError(
        'Tags may only contain letters, digits, - and _ (max. 30 characters).',
      );
    tags.push(tag);
  }
  if (tags.length > MAX_TAGS) throw new BookmarkValidationError('A bookmark can have at most 10 tags.');
  return { folderId: input.folderId.trim(), title, url, description, tags: tags.sort() };
}

function folderName(name: string): string {
  const n = name.trim();
  if (n === '' || chars(n) > MAX_FOLDER_NAME)
    throw new BookmarkValidationError('The folder name must be between 1 and 60 characters long.');
  return n;
}

const ENTITIES: Record<string, string> = { amp: '&', lt: '<', gt: '>', quot: '"', apos: "'", '#39': "'" };
const decode = (s: string) =>
  s.replace(/&(#\d+|#x[0-9a-f]+|\w+);/gi, (m, e: string) => {
    if (e[0] === '#') {
      const code = e[1] === 'x' || e[1] === 'X' ? parseInt(e.slice(2), 16) : parseInt(e.slice(1), 10);
      return Number.isFinite(code) ? String.fromCodePoint(code) : m;
    }
    return ENTITIES[e.toLowerCase()] ?? m;
  });

/**
 * Reads the links of a Netscape bookmark file with their innermost folder.
 * The format is a fixed tag soup (<DL>, <DT><H3>, <DT><A HREF>), so a small
 * tokenizer is enough and works the same in the page and in tests.
 */
export function parseNetscape(html: string): { url: string; title: string; folder: string }[] {
  const out: { url: string; title: string; folder: string }[] = [];
  const stack: string[] = [];
  let pending = '';
  const tag = /<(\/?)(dl|h3|a)\b([^>]*)>([^<]*)/gi;
  for (const m of html.matchAll(tag)) {
    const close = m[1] ?? '';
    const name = (m[2] ?? '').toLowerCase();
    const attrs = m[3] ?? '';
    const text = (m[4] ?? '').trim();
    if (name === 'h3' && !close) pending = decode(text);
    else if (name === 'dl' && !close) {
      stack.push(pending);
      pending = '';
    } else if (name === 'dl' && close) stack.pop();
    else if (name === 'a' && !close) {
      const href = /href\s*=\s*"([^"]*)"/i.exec(attrs)?.[1] ?? '';
      out.push({ url: decode(href), title: decode(text), folder: stack.at(-1) ?? '' });
    }
  }
  return out;
}

const toBookmark = (r: BookmarkRow): Bookmark => ({
  id: r.id,
  folderId: r.folderId,
  title: r.title,
  url: r.url,
  description: r.description,
  tags: [...r.tags],
  // No link checks and no stored icons without a server: the browser loads
  // the site's own favicon, as it does on the server without a stored copy.
  link: { state: '', code: 0, target: '', checkedAt: '' },
  hasIcon: false,
  createdAt: r.createdAt,
  visibility: 'private',
  mine: true,
});

/**
 * Bookmarks gateway of the browser extension's local mode (ADR 0019):
 * bookmarks and folders in one document of the DocumentStore. Fetching
 * titles and icons and checking links need requests to other sites from a
 * server, so they stay off; the import reads the browser's HTML export in
 * the page.
 */
export class LocalBookmarksGateway implements BookmarksGateway {
  constructor(
    private readonly store: DocumentStore,
    private readonly now: () => Date = () => new Date(),
  ) {}

  private async read(): Promise<BookmarksDoc> {
    return (await this.store.get<BookmarksDoc>('bookmarks')) ?? { bookmarks: [], folders: [] };
  }

  private async change<T>(fn: (doc: BookmarksDoc) => T): Promise<T> {
    const doc = await this.read();
    const result = fn(doc);
    await this.store.put('bookmarks', doc);
    return result;
  }

  private static checkFolder(doc: BookmarksDoc, id: string): void {
    if (id && !doc.folders.some((f) => f.id === id)) throw new NotFoundError('The folder does not exist.');
  }

  async list(filter: BookmarkFilter): Promise<BookmarkList> {
    const doc = await this.read();
    const q = (filter.query ?? '').trim().toLowerCase();
    const tag = (filter.tag ?? '').toLowerCase();
    // Nothing is ever checked locally, so "broken only" is always empty.
    const bookmarks = filter.broken
      ? []
      : doc.bookmarks
          .filter((b) => {
            if (filter.folderId === 'none' && b.folderId !== '') return false;
            if (filter.folderId && filter.folderId !== 'none' && b.folderId !== filter.folderId) return false;
            if (tag && !b.tags.includes(tag)) return false;
            return !q || `${b.title} ${b.url} ${b.description}`.toLowerCase().includes(q);
          })
          .sort((a, b) => a.title.localeCompare(b.title, undefined, { sensitivity: 'base' }));
    const perFolder = new Map<string, number>();
    const perTag = new Map<string, number>();
    for (const b of doc.bookmarks) {
      if (b.folderId) perFolder.set(b.folderId, (perFolder.get(b.folderId) ?? 0) + 1);
      for (const t of b.tags) perTag.set(t, (perTag.get(t) ?? 0) + 1);
    }
    const folders: Folder[] = [...doc.folders]
      .sort((a, b) => a.position - b.position || a.name.localeCompare(b.name))
      .map((f) => ({
        id: f.id,
        name: f.name,
        count: perFolder.get(f.id) ?? 0,
        visibility: 'private',
        mine: true,
      }));
    const tags = [...perTag]
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => a.name.localeCompare(b.name));
    return { bookmarks: bookmarks.map(toBookmark), folders, tags };
  }

  create(input: BookmarkInput): Promise<Bookmark> {
    return this.change((doc) => {
      const fields = normalise(input);
      LocalBookmarksGateway.checkFolder(doc, fields.folderId);
      const row: BookmarkRow = { id: newId(), ...fields, createdAt: this.now().toISOString() };
      doc.bookmarks.push(row);
      return toBookmark(row);
    });
  }

  update(id: string, input: BookmarkInput): Promise<Bookmark> {
    return this.change((doc) => {
      const row = doc.bookmarks.find((b) => b.id === id);
      if (!row) throw new NotFoundError('The bookmark does not exist.');
      const fields = normalise(input);
      LocalBookmarksGateway.checkFolder(doc, fields.folderId);
      Object.assign(row, fields);
      return toBookmark(row);
    });
  }

  delete(id: string): Promise<void> {
    return this.change((doc) => {
      doc.bookmarks = doc.bookmarks.filter((b) => b.id !== id);
    });
  }

  createFolder(name: string): Promise<Folder> {
    return this.change((doc) => {
      const row: FolderRow = { id: newId(), name: folderName(name), position: doc.folders.length };
      doc.folders.push(row);
      return { id: row.id, name: row.name, count: 0, visibility: 'private', mine: true };
    });
  }

  renameFolder(id: string, name: string): Promise<Folder> {
    return this.change((doc) => {
      const row = doc.folders.find((f) => f.id === id);
      if (!row) throw new NotFoundError('The folder does not exist.');
      row.name = folderName(name);
      const count = doc.bookmarks.filter((b) => b.folderId === id).length;
      return { id: row.id, name: row.name, count, visibility: 'private', mine: true };
    });
  }

  /** Deletes a folder; its bookmarks are kept without folder, as on the server. */
  deleteFolder(id: string): Promise<void> {
    return this.change((doc) => {
      doc.folders = doc.folders.filter((f) => f.id !== id);
      for (const b of doc.bookmarks) if (b.folderId === id) b.folderId = '';
    });
  }

  async shareFolder(id: string): Promise<Folder> {
    // One local account: nobody to share with, the folder stays private.
    const { folders } = await this.list({});
    const f = folders.find((x) => x.id === id);
    if (!f) throw new NotFoundError('The folder does not exist.');
    return f;
  }

  shareGroups(): Promise<ShareGroup[]> {
    return Promise.resolve([]);
  }

  /**
   * Imports a browser's bookmark export (Netscape HTML). Folders are the
   * <H3> headings; nested folders are flattened to the innermost name, and
   * links already stored are skipped, as the server's import does.
   */
  importHtml(html: string): Promise<BookmarkImportResult> {
    const result: BookmarkImportResult = {
      created: 0,
      skippedInvalid: 0,
      skippedExists: 0,
      foldersCreated: 0,
    };
    return this.change((doc) => {
      const known = new Set(doc.bookmarks.map((b) => b.url));
      for (const link of parseNetscape(html)) {
        let fields: Omit<BookmarkRow, 'id' | 'createdAt'>;
        try {
          fields = normalise({ folderId: '', title: link.title, url: link.url, description: '', tags: [] });
        } catch {
          result.skippedInvalid++;
          continue;
        }
        if (known.has(fields.url)) {
          result.skippedExists++;
          continue;
        }
        if (link.folder) {
          const name = link.folder.slice(0, MAX_FOLDER_NAME);
          let row = doc.folders.find((f) => f.name === name);
          if (!row) {
            row = { id: newId(), name, position: doc.folders.length };
            doc.folders.push(row);
            result.foldersCreated++;
          }
          fields.folderId = row.id;
        }
        doc.bookmarks.push({ id: newId(), ...fields, createdAt: this.now().toISOString() });
        known.add(fields.url);
        result.created++;
      }
      return result;
    });
  }

  getSettings(): Promise<BookmarkSettings> {
    return Promise.resolve({ enrich: false, checkLinks: false });
  }

  /** Both features need a server; the switches stay off. */
  saveSettings(): Promise<BookmarkSettings> {
    return this.getSettings();
  }

  async enrich(id: string): Promise<Bookmark> {
    const row = (await this.read()).bookmarks.find((b) => b.id === id);
    if (!row) throw new NotFoundError('The bookmark does not exist.');
    return toBookmark(row);
  }

  checkLink(id: string): Promise<Bookmark> {
    return this.enrich(id);
  }
}
