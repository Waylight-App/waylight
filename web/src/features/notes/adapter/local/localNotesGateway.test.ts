import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import { plain } from '../../domain/note';
import { LocalNotesGateway, diffLines } from './localNotesGateway';

describe('LocalNotesGateway', () => {
  it('saves, lists pinned first, searches and links notes', async () => {
    const gw = new LocalNotesGateway(new MemoryDocumentStore());
    const shop = await gw.save({ title: 'Shopping', body: 'milk, bread', pinned: false });
    const plan = await gw.save({ title: 'Week plan', body: 'Monday: see [[shopping]]', pinned: true });
    expect((await gw.list(false)).map((n) => n.title)).toEqual(['Week plan', 'Shopping']);
    expect((await gw.list(true)).map((n) => n.id)).toEqual([plan.id]);
    const hits = await gw.search('BREAD');
    expect(hits.map((h) => h.id)).toEqual([shop.id]);
    expect(plain(hits[0]?.snippet ?? [])).toContain('bread');
    expect(hits[0]?.snippet.some((s) => s.hit)).toBe(true);
    expect(await gw.backlinks(shop.id)).toEqual([{ id: plan.id, title: 'Week plan' }]);
  });

  it('keeps one revision per ten minutes of editing and restores it', async () => {
    let t = new Date('2026-11-16T08:00:00Z');
    const gw = new LocalNotesGateway(new MemoryDocumentStore(), () => t);
    const n = await gw.save({ title: 'Draft', body: 'one', pinned: false });
    await gw.save({ title: 'Draft', body: 'two', pinned: false }, n.id);
    t = new Date('2026-11-16T08:05:00Z');
    await gw.save({ title: 'Draft', body: 'three', pinned: false }, n.id);
    let revs = await gw.revisions(n.id);
    expect(revs).toHaveLength(1);
    t = new Date('2026-11-16T08:20:00Z');
    await gw.save({ title: 'Draft', body: 'four', pinned: false }, n.id);
    revs = await gw.revisions(n.id);
    expect(revs).toHaveLength(2);
    const oldest = revs[revs.length - 1]?.id ?? '';
    const detail = await gw.revision(n.id, oldest);
    expect(detail.body).toBe('one');
    expect(detail.diff).toEqual([
      { op: '-', text: 'one' },
      { op: '+', text: 'four' },
    ]);
    expect((await gw.restore(n.id, oldest)).body).toBe('one');
    await gw.delete(n.id);
    expect(await gw.list(false)).toEqual([]);
  });

  it('diffs lines like the server', () => {
    expect(diffLines('a\nb\nc', 'a\nc\nd')).toEqual([
      { op: ' ', text: 'a' },
      { op: '-', text: 'b' },
      { op: ' ', text: 'c' },
      { op: '+', text: 'd' },
    ]);
  });
});
