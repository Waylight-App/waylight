import { MarketsApi } from '../../features/markets/adapter/api/marketsApi';
import { MarketsUseCases } from '../../features/markets/usecase/marketsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    markets: MarketsUseCases;
  }
}

export const service = defineService('markets', (http) => new MarketsUseCases(new MarketsApi(http)));
