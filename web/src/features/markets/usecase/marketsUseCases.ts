import type { MarketRow } from '../domain/markets';

/** Output port: the markets API. */
export interface MarketsGateway {
  watchlist(symbols: string[]): Promise<MarketRow[]>;
}

/** The watchlist of a widget (#284). */
export class MarketsUseCases {
  constructor(private readonly gateway: MarketsGateway) {}

  watchlist(symbols: string[]): Promise<MarketRow[]> {
    return this.gateway.watchlist(symbols);
  }
}
