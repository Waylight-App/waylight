import type { Rates } from '../../search/domain/calc/currency';
import type { MarketRow } from '../domain/markets';

/** Output port: the markets API. */
export interface MarketsGateway {
  watchlist(symbols: string[]): Promise<MarketRow[]>;
  /** The ECB's daily reference rates for the search (#482). */
  rates(): Promise<Rates>;
}

/** The watchlist of a widget (#284). */
export class MarketsUseCases {
  constructor(private readonly gateway: MarketsGateway) {}

  watchlist(symbols: string[]): Promise<MarketRow[]> {
    return this.gateway.watchlist(symbols);
  }

  rates(): Promise<Rates> {
    return this.gateway.rates();
  }
}
