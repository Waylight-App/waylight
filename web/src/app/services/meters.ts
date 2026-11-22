import { MetersApi } from '../../features/meters/adapter/api/metersApi';
import { MetersUseCases } from '../../features/meters/usecase/metersUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    meters: MetersUseCases;
  }
}

export const service = defineService('meters', (http) => new MetersUseCases(new MetersApi(http)));
