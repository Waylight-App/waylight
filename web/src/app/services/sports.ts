import { SportsApi } from '../../features/sports/adapter/api/sportsApi';
import { SportsUseCases } from '../../features/sports/usecase/sportsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    sports: SportsUseCases;
  }
}

export const service = defineService('sports', (http) => new SportsUseCases(new SportsApi(http)));
