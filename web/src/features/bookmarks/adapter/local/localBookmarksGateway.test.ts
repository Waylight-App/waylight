import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import { BookmarksUseCases } from '../../usecase/bookmarksUseCases';
import { LocalBookmarksGateway } from './localBookmarksGateway';

const input = (url: string, more: Partial<{ title: string; folderId: string; tags: string[] }> = {}) => ({
  url,
  title: more.title ?? '',
  description: '',
  folderId: more.folderId ?? '',
  tags: more.tags ?? [],
});

describe('LocalBookmarksGateway', () => {
  it('validates like the server and fills the title from the host', async () => {
    const uc = new BookmarksUseCases(new LocalBookmarksGateway(new MemoryDocumentStore()));
    const b = await uc.save(input('example.org/docs', { tags: ['#Docs', 'docs', 'Work'] }));
    expect(b.url).toBe('https://example.org/docs');
    expect(b.title).toBe('example.org');
    expect(b.tags).toEqual(['docs', 'work']);
    await expect(uc.save(input('ftp://example.org'))).rejects.toThrow('http://');
    await expect(uc.save(input('https://example.org', { tags: ['a b'] }))).rejects.toThrow('Tags');
  });

  it('files bookmarks into folders, filters and counts them', async () => {
    const uc = new BookmarksUseCases(new LocalBookmarksGateway(new MemoryDocumentStore()));
    const work = await uc.createFolder(' Work ');
    await uc.save(input('https://b.example', { title: 'Beta', folderId: work.id, tags: ['x'] }));
    await uc.save(input('https://a.example', { title: 'alpha' }));
    const all = await uc.list();
    expect(all.bookmarks.map((b) => b.title)).toEqual(['alpha', 'Beta']);
    expect(all.folders).toEqual([{ id: work.id, name: 'Work', count: 1, visibility: 'private', mine: true }]);
    expect(all.tags).toEqual([{ name: 'x', count: 1 }]);
    expect((await uc.list({ folderId: 'none' })).bookmarks.map((b) => b.title)).toEqual(['alpha']);
    expect((await uc.list({ query: 'BETA' })).bookmarks).toHaveLength(1);
    await uc.deleteFolder(work.id);
    expect((await uc.list({ folderId: 'none' })).bookmarks).toHaveLength(2);
    await expect(uc.save(input('https://c.example', { folderId: 'missing' }))).rejects.toThrow('folder');
  });

  it('imports a browser export and skips links it already has', async () => {
    const gw = new LocalBookmarksGateway(new MemoryDocumentStore());
    const html = `<!DOCTYPE NETSCAPE-Bookmark-file-1><DL><p>
      <DT><H3>Tools</H3><DL><p>
        <DT><A HREF="https://tool.example">Tool</A>
        <DT><A HREF="javascript:alert(1)">Bad</A>
      </DL><p>
      <DT><A HREF="https://news.example">News</A>
    </DL>`;
    expect(await gw.importHtml(html)).toEqual({
      created: 2,
      skippedInvalid: 1,
      skippedExists: 0,
      foldersCreated: 1,
    });
    expect((await gw.importHtml(html)).skippedExists).toBe(2);
    const list = await gw.list({});
    expect(list.folders.map((f) => f.name)).toEqual(['Tools']);
    expect(list.bookmarks.find((b) => b.title === 'Tool')?.folderId).toBe(list.folders[0]?.id);
  });

  it('keeps the server-only features off', async () => {
    const gw = new LocalBookmarksGateway(new MemoryDocumentStore());
    expect(await gw.saveSettings()).toEqual({ enrich: false, checkLinks: false });
    expect((await gw.list({ broken: true })).bookmarks).toEqual([]);
  });
});
