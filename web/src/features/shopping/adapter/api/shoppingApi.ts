import type { HttpClient } from '../../../../shared/api-client/http';
import type { ClearedDTO, ItemDTO, SettingsDTO } from '../../../../shared/api-client/generated/shopping';
import type {
  ItemStateDTO,
  NewItemDTO,
  SettingsInputDTO,
} from '../../../../shared/api-client/generated/shopping-requests';
import { fetchShareGroups, type ShareGroup, type Visibility } from '../../../../shared/sharing/sharing';
import type { ShoppingItem, ShoppingList, ShoppingOverview, ShoppingSettings } from '../../domain/shopping';
import type { ShoppingGateway } from '../../usecase/shoppingUseCases';

// TODO(#471): replace with the generated DTOs once the backend lands.
interface ListDTO {
  id: string;
  name: string;
  visibility: string;
  mine: boolean;
  open: number;
}
interface ListsDTO {
  lists: ListDTO[];
  caldav: boolean;
  current: string;
}

const item = (id: string) => `/api/shopping/items/${encodeURIComponent(id)}`;
const list = (id: string) => `/api/shopping/lists/${encodeURIComponent(id)}`;
const listItem = (listId: string, id: string) => `${list(listId)}/items/${encodeURIComponent(id)}`;

/** ShoppingGateway backed by the REST API. */
export class ShoppingApi implements ShoppingGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<ShoppingSettings> {
    return { ...(await this.http.get<SettingsDTO>('/api/shopping/settings')) };
  }
  async saveSettings(listUrl: string, username: string, password: string): Promise<ShoppingSettings> {
    const body: SettingsInputDTO = { listUrl, username, password };
    return { ...(await this.http.put<SettingsDTO>('/api/shopping/settings', body)) };
  }
  async items(): Promise<ShoppingItem[]> {
    return (await this.http.get<ItemDTO[]>('/api/shopping/items')).map((i) => ({ ...i }));
  }
  async add(summary: string): Promise<ShoppingItem> {
    const body: NewItemDTO = { summary };
    return { ...(await this.http.post<ItemDTO>('/api/shopping/items', body)) };
  }
  setDone(id: string, done: boolean): Promise<void> {
    const body: ItemStateDTO = { done };
    return this.http.put(item(id), body);
  }
  remove(id: string): Promise<void> {
    return this.http.delete(item(id));
  }
  async clearCompleted(): Promise<number> {
    return (await this.http.post<ClearedDTO>('/api/shopping/clear-completed')).deleted;
  }

  async overview(): Promise<ShoppingOverview> {
    const o = await this.http.get<ListsDTO>('/api/shopping/lists');
    return { lists: o.lists.map((l) => ({ ...l })), caldav: o.caldav, current: o.current };
  }
  async createList(name: string, visibility?: Visibility): Promise<ShoppingList> {
    return { ...(await this.http.post<ListDTO>('/api/shopping/lists', { name, visibility })) };
  }
  async updateList(id: string, name: string, visibility: Visibility): Promise<ShoppingList> {
    return { ...(await this.http.put<ListDTO>(list(id), { name, visibility })) };
  }
  deleteList(id: string): Promise<void> {
    return this.http.delete(list(id));
  }
  setCurrent(current: string): Promise<void> {
    return this.http.put('/api/shopping/current', { list: current });
  }
  async listItems(listId: string): Promise<ShoppingItem[]> {
    return (await this.http.get<ItemDTO[]>(`${list(listId)}/items`)).map((i) => ({ ...i }));
  }
  async addToList(listId: string, summary: string): Promise<ShoppingItem> {
    const body: NewItemDTO = { summary };
    return { ...(await this.http.post<ItemDTO>(`${list(listId)}/items`, body)) };
  }
  setDoneInList(listId: string, id: string, done: boolean): Promise<void> {
    const body: ItemStateDTO = { done };
    return this.http.put(listItem(listId, id), body);
  }
  removeFromList(listId: string, id: string): Promise<void> {
    return this.http.delete(listItem(listId, id));
  }
  async clearCompletedInList(listId: string): Promise<number> {
    return (await this.http.post<ClearedDTO>(`${list(listId)}/clear-completed`)).deleted;
  }
  async importCaldav(listId: string): Promise<number> {
    return (await this.http.post<{ imported: number }>(`${list(listId)}/import`)).imported;
  }
  shareGroups(): Promise<ShareGroup[]> {
    return fetchShareGroups(this.http);
  }
}
