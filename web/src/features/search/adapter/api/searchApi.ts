import type { HttpClient } from '../../../../shared/api-client/http';
import type { KeywordDTO, SettingsDTO, ShortcutDTO } from '../../../../shared/api-client/generated/search';
import { KEYWORD_KINDS, type KeywordKind } from '../../domain/keywords';
import type {
  BarOptionsDTO,
  DefaultProviderDTO,
} from '../../../../shared/api-client/generated/search-requests';
import type { SearchBarOptions, SearchGateway, SearchSettings, Shortcut } from '../../usecase/ports';
import type { OwnKeyword } from '../../domain/keywords';

const toSettings = (dto: SettingsDTO): SearchSettings => ({
  defaultProvider: dto.defaultProvider,
  providers: dto.providers.map((p) => ({ id: p.id, name: p.name })),
  shortcuts: dto.shortcuts.map((s) => ({ keyword: s.keyword, name: s.name, urlTemplate: s.urlTemplate })),
  // A kind this client does not know (a newer server) is left out.
  keywords: dto.keywords
    .filter((k) => (KEYWORD_KINDS as readonly string[]).includes(k.kind))
    .map((k) => ({ keyword: k.keyword, kind: k.kind as KeywordKind, target: k.target, label: k.label })),
  bar: {
    openInNewTab: dto.bar.openInNewTab,
    capturePrefixes: { ...dto.bar.capturePrefixes },
  },
});

/** SearchGateway backed by the REST API. */
export class SearchApi implements SearchGateway {
  constructor(private readonly http: HttpClient) {}

  async settings(): Promise<SearchSettings> {
    return toSettings(await this.http.get<SettingsDTO>('/api/search/settings'));
  }

  async setDefaultProvider(id: string): Promise<SearchSettings> {
    const body: DefaultProviderDTO = { defaultProvider: id };
    return toSettings(await this.http.put<SettingsDTO>('/api/search/settings', body));
  }

  async setBar(options: SearchBarOptions): Promise<SearchSettings> {
    const body: BarOptionsDTO = {
      openInNewTab: options.openInNewTab,
      capturePrefixes: { ...options.capturePrefixes },
    };
    return toSettings(await this.http.put<SettingsDTO>('/api/search/bar', body));
  }

  async addShortcut(s: Shortcut): Promise<SearchSettings> {
    const body: ShortcutDTO = { keyword: s.keyword, name: s.name, urlTemplate: s.urlTemplate };
    return toSettings(await this.http.post<SettingsDTO>('/api/search/shortcuts', body));
  }

  async deleteShortcut(keyword: string): Promise<SearchSettings> {
    return toSettings(
      await this.http.delete<SettingsDTO>(`/api/search/shortcuts/${encodeURIComponent(keyword)}`),
    );
  }

  async addKeyword(k: OwnKeyword): Promise<SearchSettings> {
    const body: KeywordDTO = { keyword: k.keyword, kind: k.kind, target: k.target, label: k.label };
    return toSettings(await this.http.post<SettingsDTO>('/api/search/keywords', body));
  }

  async deleteKeyword(keyword: string): Promise<SearchSettings> {
    return toSettings(
      await this.http.delete<SettingsDTO>(`/api/search/keywords/${encodeURIComponent(keyword)}`),
    );
  }

  searchUrl(query: string): string {
    return `/api/search?q=${encodeURIComponent(query)}`;
  }
}
