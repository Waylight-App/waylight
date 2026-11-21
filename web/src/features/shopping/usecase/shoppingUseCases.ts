import type { ShareGroup, Visibility } from '../../../shared/sharing/sharing';
import {
  CALDAV,
  type ShoppingItem,
  type ShoppingList,
  type ShoppingOverview,
  type ShoppingSettings,
} from '../domain/shopping';

/** Output port: shopping list backend. */
export interface ShoppingGateway {
  // The connected CalDAV task list.
  settings(): Promise<ShoppingSettings>;
  /** signIn is a finished Nextcloud sign-in of the setup assistant (#654) that replaces the credential. */
  saveSettings(
    listUrl: string,
    username: string,
    password: string,
    signIn?: string,
  ): Promise<ShoppingSettings>;
  items(): Promise<ShoppingItem[]>;
  add(summary: string): Promise<ShoppingItem>;
  setDone(id: string, done: boolean): Promise<void>;
  remove(id: string): Promise<void>;
  clearCompleted(): Promise<number>;

  // Lists kept on Waylight (#471).
  overview(): Promise<ShoppingOverview>;
  createList(name: string, visibility?: Visibility): Promise<ShoppingList>;
  updateList(id: string, name: string, visibility: Visibility): Promise<ShoppingList>;
  deleteList(id: string): Promise<void>;
  setCurrent(list: string): Promise<void>;
  listItems(listId: string): Promise<ShoppingItem[]>;
  addToList(listId: string, summary: string): Promise<ShoppingItem>;
  setDoneInList(listId: string, id: string, done: boolean): Promise<void>;
  removeFromList(listId: string, id: string): Promise<void>;
  clearCompletedInList(listId: string): Promise<number>;
  importCaldav(listId: string): Promise<number>;
  shareGroups(): Promise<ShareGroup[]>;
}

/**
 * The items of the list the widget shows, whichever kind it is. The widget
 * renders both kinds the same way, so it only talks to this.
 */
export interface ItemSource {
  items(): Promise<ShoppingItem[]>;
  /** Adds an item; empty input is ignored. */
  add(summary: string): Promise<ShoppingItem | null>;
  setDone(id: string, done: boolean): Promise<void>;
  remove(id: string): Promise<void>;
  clearCompleted(): Promise<number>;
}

/** Application logic of the shopping list feature. */
export class ShoppingUseCases {
  constructor(private readonly gateway: ShoppingGateway) {}

  settings(): Promise<ShoppingSettings> {
    return this.gateway.settings();
  }
  saveSettings(listUrl: string, username: string, password: string): Promise<ShoppingSettings> {
    return this.gateway.saveSettings(listUrl.trim(), username.trim(), password);
  }
  /** Saves the list with a finished Nextcloud sign-in of the setup assistant (#654). */
  saveSignIn(listUrl: string, signIn: string): Promise<ShoppingSettings> {
    return this.gateway.saveSettings(listUrl.trim(), '', '', signIn);
  }

  overview(): Promise<ShoppingOverview> {
    return this.gateway.overview();
  }
  /** Creates a list; an empty name is refused before asking the server. */
  async createList(name: string, visibility?: Visibility): Promise<ShoppingList | null> {
    const n = name.trim();
    return n ? this.gateway.createList(n, visibility) : null;
  }
  updateList(id: string, name: string, visibility: Visibility): Promise<ShoppingList> {
    return this.gateway.updateList(id, name.trim(), visibility);
  }
  deleteList(id: string): Promise<void> {
    return this.gateway.deleteList(id);
  }
  /** Remembers which list the widget shows for this account. */
  setCurrent(list: string): Promise<void> {
    return this.gateway.setCurrent(list);
  }
  /** Copies the open items of the CalDAV list into a list; returns how many. */
  importCaldav(listId: string): Promise<number> {
    return this.gateway.importCaldav(listId);
  }
  shareGroups(): Promise<ShareGroup[]> {
    return this.gateway.shareGroups();
  }

  /** The items of the CalDAV list (CALDAV) or of a list on Waylight. */
  source(current: string): ItemSource {
    const g = this.gateway;
    const trimmed = async (summary: string, add: (s: string) => Promise<ShoppingItem>) => {
      const s = summary.trim();
      return s ? add(s) : null;
    };
    if (current === CALDAV) {
      return {
        items: () => g.items(),
        add: (s) => trimmed(s, (x) => g.add(x)),
        setDone: (id, done) => g.setDone(id, done),
        remove: (id) => g.remove(id),
        clearCompleted: () => g.clearCompleted(),
      };
    }
    return {
      items: () => g.listItems(current),
      add: (s) => trimmed(s, (x) => g.addToList(current, x)),
      setDone: (id, done) => g.setDoneInList(current, id, done),
      remove: (id) => g.removeFromList(current, id),
      clearCompleted: () => g.clearCompletedInList(current),
    };
  }
}
