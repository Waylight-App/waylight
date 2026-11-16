import { EXPORT_OWNER, exportTime, type ExportTables } from '../../../../shared/local/exportRows';
import type { BookmarksDoc } from './localBookmarksGateway';

/**
 * The bookmarks document as rows of bookmark_folders, bookmarks and
 * bookmark_tags (migrations of internal/modules/bookmarks), parents first.
 * Everything belongs to the owner and stays private; link checks and icons
 * do not exist locally.
 */
export function bookmarksExport(doc: BookmarksDoc): ExportTables {
  return {
    bookmark_folders: doc.folders.map((f) => ({
      id: f.id,
      name: f.name,
      position: f.position,
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
    bookmarks: doc.bookmarks.map((b) => ({
      id: b.id,
      // NULL = unfiled, as the server's foreign key expects.
      folder_id: b.folderId || null,
      title: b.title,
      url: b.url,
      description: b.description,
      created_at: exportTime(b.createdAt),
      owner_id: EXPORT_OWNER,
      visibility: 'private',
    })),
    bookmark_tags: doc.bookmarks.flatMap((b) => b.tags.map((tag) => ({ bookmark_id: b.id, tag }))),
  };
}
