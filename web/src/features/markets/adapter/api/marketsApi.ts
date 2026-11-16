import type { HttpClient } from '../../../../shared/api-client/http';
import type { RatesDTO, RowDTO } from '../../../../shared/api-client/generated/markets';
import type { Rates } from '../../../search/domain/calc/currency';
import type { MarketRow } from '../../domain/markets';
import type { MarketsGateway } from '../../usecase/marketsUseCases';

/** MarketsGateway backed by the REST API. */
export class MarketsApi implements MarketsGateway {
  constructor(private readonly http: HttpClient) {}

  async watchlist(symbols: string[]): Promise<MarketRow[]> {
    const rows = await this.http.get<RowDTO[]>(
      `/api/markets/watchlist?symbols=${encodeURIComponent(symbols.join(','))}`,
    );
    return rows.map((r) => ({
      symbol: r.symbol,
      value: r.value ?? null,
      changePct: r.changePct,
      asOf: r.asOf,
      history: (r.history ?? []).map((p) => p.value),
      stale: r.stale,
      error: r.error ?? '',
    }));
  }

  async rates(): Promise<Rates> {
    const r = await this.http.get<RatesDTO>('/api/markets/rates');
    return { base: r.base, date: r.date, rates: r.rates, stale: r.stale };
  }
}
