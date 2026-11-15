import { describe, expect, it } from 'vitest';
import { MemoryDocumentStore } from '../../../../shared/local/documentStore';
import { LocalDashboardGateway } from './localDashboardGateway';

/** The first element; fails the test instead of a non-null assertion. */
function first<T>(items: readonly T[]): T {
  const item = items[0];
  if (item === undefined) throw new Error('expected at least one item');
  return item;
}

const setup = () => {
  const store = new MemoryDocumentStore();
  return { store, gateway: new LocalDashboardGateway(store, () => new Date('2026-11-16T12:00:00')) };
};

describe('LocalDashboardGateway', () => {
  it('creates a Home tab with local widgets on first use, once', async () => {
    const { gateway } = setup();
    const first = await gateway.load();
    expect(first.tabs.map((t) => t.name)).toEqual(['Home']);
    expect(first.tabs[0]?.widgets.map((w) => w.kind)).toEqual(['clock', 'todos', 'notes', 'bookmarks']);
    expect((await gateway.load()).tabs).toHaveLength(1);
  });

  it('keeps tabs and widgets across gateway instances on the same store', async () => {
    const { store, gateway } = setup();
    const tab = await gateway.createTab('Work');
    await gateway.addWidget(tab.id, 'focus', { x: 0, y: 0, w: 3, h: 3 }, { minutes: 25 });
    const again = await new LocalDashboardGateway(store).load();
    const work = again.tabs.find((t) => t.name === 'Work');
    expect(work?.widgets).toEqual([expect.objectContaining({ kind: 'focus', config: { minutes: 25 } })]);
  });

  it('orders tabs by position and reorders them', async () => {
    const { gateway } = setup();
    const home = first((await gateway.load()).tabs);
    const work = await gateway.createTab('Work');
    expect(work.position).toBe(1);
    await gateway.reorderTabs([work.id, home.id]);
    expect((await gateway.load()).tabs.map((t) => t.name)).toEqual(['Work', 'Home']);
  });

  it('deletes a tab together with its widgets', async () => {
    const { store, gateway } = setup();
    const tab = await gateway.createTab('Temp');
    await gateway.addWidget(tab.id, 'notes', { x: 0, y: 0, w: 3, h: 3 }, {});
    await gateway.deleteTab(tab.id);
    const doc = await store.get<{ widgets: { tabId: string }[] }>('dashboard');
    expect(doc?.widgets.some((w) => w.tabId === tab.id)).toBe(false);
  });

  it('saves the layout and rejects widgets of another tab', async () => {
    const { gateway } = setup();
    const home = first((await gateway.load()).tabs);
    const clock = first(home.widgets);
    await gateway.saveLayout(home.id, [{ id: clock.id, x: 9, y: 9, w: 3, h: 2 }]);
    const moved = first((await gateway.load()).tabs).widgets.find((w) => w.id === clock.id);
    expect(moved).toMatchObject({ x: 9, y: 9, w: 3, h: 2 });
    const other = await gateway.createTab('Other');
    await expect(gateway.saveLayout(other.id, [{ id: clock.id, x: 0, y: 0, w: 1, h: 1 }])).rejects.toThrow();
  });

  it('follows the desktop grid on phones until a phone layout is saved, and back after reset', async () => {
    const { gateway } = setup();
    const home = first((await gateway.load()).tabs);
    expect(home.mobileCustom).toBe(false);
    expect(home.mobile).toHaveLength(4);
    const custom = await gateway.saveMobileLayout(home.id, [
      { id: first(home.widgets).id, hidden: true, rows: 2 },
    ]);
    expect(custom.mobileCustom).toBe(true);
    expect(first(custom.mobile)).toEqual({ id: first(home.widgets).id, hidden: true, rows: 2 });
    expect((await gateway.resetMobileLayout(home.id)).mobileCustom).toBe(false);
  });

  it('saves a template and creates a tab from it with new widget ids', async () => {
    const { gateway } = setup();
    const home = first((await gateway.load()).tabs);
    const template = await gateway.saveTemplate(home.id, 'Start');
    expect(template).toMatchObject({ name: 'Start', widgets: 4 });
    expect(await gateway.listTemplates()).toHaveLength(1);
    const copy = await gateway.createTabFromTemplate(template.id, 'Copy');
    expect(copy.widgets.map((w) => w.kind)).toEqual(home.widgets.map((w) => w.kind));
    expect(copy.widgets.some((w) => home.widgets.some((h) => h.id === w.id))).toBe(false);
    await gateway.deleteTemplate(template.id);
    expect(await gateway.listTemplates()).toEqual([]);
  });

  it('picks the default tab from its rules', async () => {
    const { gateway } = setup();
    const work = await gateway.createTab('Work');
    await gateway.setRules(work.id, [{ days: [1], from: '', to: '', networks: [] }]);
    expect((await gateway.load()).defaultTabId).toBe(work.id);
  });
});
