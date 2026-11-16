import { describe, expect, it, vi } from 'vitest';
import { ShoppingUseCases, type ShoppingGateway } from './shoppingUseCases';

const item = { id: 'a', summary: 'Milk', done: false };

function fakeGateway() {
  return {
    settings: vi.fn(),
    saveSettings: vi.fn(),
    items: vi.fn().mockResolvedValue([item]),
    add: vi.fn().mockResolvedValue(item),
    setDone: vi.fn(),
    remove: vi.fn(),
    clearCompleted: vi.fn(),
    overview: vi.fn(),
    createList: vi
      .fn()
      .mockResolvedValue({ id: 'l', name: 'Groceries', visibility: 'private', mine: true, open: 0 }),
    updateList: vi.fn(),
    deleteList: vi.fn(),
    setCurrent: vi.fn(),
    listItems: vi.fn().mockResolvedValue([]),
    addToList: vi.fn().mockResolvedValue(item),
    setDoneInList: vi.fn(),
    removeFromList: vi.fn(),
    clearCompletedInList: vi.fn(),
    importCaldav: vi.fn(),
    shareGroups: vi.fn(),
  } satisfies ShoppingGateway;
}

describe('ShoppingUseCases', () => {
  it('sends CalDAV items to the CalDAV routes, ignoring empty input', async () => {
    const g = fakeGateway();
    const src = new ShoppingUseCases(g).source('caldav');
    expect(await src.add('   ')).toBeNull();
    await src.add(' Milk ');
    await src.setDone('a', true);
    expect(g.add).toHaveBeenCalledWith('Milk');
    expect(g.setDone).toHaveBeenCalledWith('a', true);
    expect(g.addToList).not.toHaveBeenCalled();
  });

  it('sends items of a list on trisse-home to that list', async () => {
    const g = fakeGateway();
    const src = new ShoppingUseCases(g).source('l1');
    await src.items();
    await src.add(' Bread ');
    await src.remove('x');
    await src.clearCompleted();
    expect(g.listItems).toHaveBeenCalledWith('l1');
    expect(g.addToList).toHaveBeenCalledWith('l1', 'Bread');
    expect(g.removeFromList).toHaveBeenCalledWith('l1', 'x');
    expect(g.clearCompletedInList).toHaveBeenCalledWith('l1');
    expect(g.items).not.toHaveBeenCalled();
  });

  it('refuses a list without a name', async () => {
    const g = fakeGateway();
    const uc = new ShoppingUseCases(g);
    expect(await uc.createList('  ')).toBeNull();
    await uc.createList(' Groceries ', 'everyone');
    expect(g.createList).toHaveBeenCalledWith('Groceries', 'everyone');
  });
});
