import { MIN_QUERY_LENGTH, type Wiki, type WikiInput, type WikiSearch } from '../domain/wiki';

/** Output port: the wiki search API. */
export interface WikiGateway {
  search(query: string): Promise<WikiSearch>;
  list(): Promise<Wiki[]>;
  create(input: WikiInput): Promise<Wiki>;
  update(id: string, input: WikiInput): Promise<Wiki>;
  remove(id: string): Promise<void>;
  test(id: string): Promise<void>;
}

/** Searching connected wikis and managing the connections (#321). */
export class WikiUseCases {
  constructor(private readonly gateway: WikiGateway) {}

  /** Asks the wikis; too short a query asks nobody. */
  search(query: string): Promise<WikiSearch> {
    const q = query.trim();
    if (q.length < MIN_QUERY_LENGTH) return Promise.resolve({ pages: [], failed: [] });
    return this.gateway.search(q);
  }
  list(): Promise<Wiki[]> {
    return this.gateway.list();
  }
  save(id: string | null, input: WikiInput): Promise<Wiki> {
    const clean = { ...input, name: input.name.trim(), url: input.url.trim(), token: input.token.trim() };
    return id ? this.gateway.update(id, clean) : this.gateway.create(clean);
  }
  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
  test(id: string): Promise<void> {
    return this.gateway.test(id);
  }
}
