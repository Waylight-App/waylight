import type { League, SportsOverview, SportsTeam } from '../domain/sports';

/** Output port: the sports API. */
export interface SportsGateway {
  leagues(): Promise<League[]>;
  teams(league: string): Promise<SportsTeam[]>;
  overview(league: string, team: number): Promise<SportsOverview>;
}

/** Results and fixtures for the widget (#285). */
export class SportsUseCases {
  constructor(private readonly gateway: SportsGateway) {}

  leagues(): Promise<League[]> {
    return this.gateway.leagues();
  }
  teams(league: string): Promise<SportsTeam[]> {
    return this.gateway.teams(league);
  }
  overview(league: string, team: number): Promise<SportsOverview> {
    return this.gateway.overview(league, team);
  }
}
