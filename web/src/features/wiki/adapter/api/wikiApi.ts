import type { HttpClient } from '../../../../shared/api-client/http';
import type { SearchDTO, WikiDTO } from '../../../../shared/api-client/generated/wiki';
import type { WikiInputDTO } from '../../../../shared/api-client/generated/wiki-requests';
import type { Wiki, WikiInput, WikiKind, WikiSearch } from '../../domain/wiki';
import type { WikiGateway } from '../../usecase/wikiUseCases';

const toWiki = (d: WikiDTO): Wiki => ({ ...d, kind: d.kind as WikiKind });

/** WikiGateway backed by the REST API. */
export class WikiApi implements WikiGateway {
  constructor(private readonly http: HttpClient) {}

  async search(query: string): Promise<WikiSearch> {
    const d = await this.http.get<SearchDTO>(`/api/wiki/search?q=${encodeURIComponent(query)}`);
    return {
      pages: d.pages.map((p) => ({ ...p, updated: p.updated ? new Date(p.updated) : null })),
      failed: d.failed ?? [],
    };
  }
  async list(): Promise<Wiki[]> {
    return (await this.http.get<WikiDTO[]>('/api/wiki/wikis')).map(toWiki);
  }
  async create(input: WikiInput): Promise<Wiki> {
    const body: WikiInputDTO = { ...input };
    return toWiki(await this.http.post<WikiDTO>('/api/wiki/wikis', body));
  }
  async update(id: string, input: WikiInput): Promise<Wiki> {
    const body: WikiInputDTO = { ...input };
    return toWiki(await this.http.put<WikiDTO>(`/api/wiki/wikis/${encodeURIComponent(id)}`, body));
  }
  async remove(id: string): Promise<void> {
    await this.http.delete(`/api/wiki/wikis/${encodeURIComponent(id)}`);
  }
  async test(id: string): Promise<void> {
    await this.http.post(`/api/wiki/wikis/${encodeURIComponent(id)}/test`);
  }
}
