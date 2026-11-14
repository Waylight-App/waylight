import type { HttpClient } from '../../../../shared/api-client/http';
import type {
  LeagueDTO,
  MatchDTO,
  OverviewDTO,
  TeamDTO,
} from '../../../../shared/api-client/generated/sports';
import type { League, SportsMatch, SportsOverview, SportsTeam } from '../../domain/sports';
import type { SportsGateway } from '../../usecase/sportsUseCases';

const toMatch = (m: MatchDTO | null | undefined): SportsMatch | null =>
  m ? { ...m, kickoff: new Date(m.kickoff) } : null;

/** SportsGateway backed by the REST API. */
export class SportsApi implements SportsGateway {
  constructor(private readonly http: HttpClient) {}

  async leagues(): Promise<League[]> {
    return (await this.http.get<LeagueDTO[]>('/api/sports/leagues')).map((l) => ({ ...l }));
  }
  async teams(league: string): Promise<SportsTeam[]> {
    return (await this.http.get<TeamDTO[]>(`/api/sports/teams?league=${encodeURIComponent(league)}`)).map(
      (t) => ({ ...t }),
    );
  }
  async overview(league: string, team: number): Promise<SportsOverview> {
    const q = `league=${encodeURIComponent(league)}${team > 0 ? `&team=${team}` : ''}`;
    const d = await this.http.get<OverviewDTO>(`/api/sports/overview?${q}`);
    return {
      league: { ...d.league },
      season: d.season,
      team: d.team ?? null,
      last: toMatch(d.last),
      next: toMatch(d.next),
      position: d.position ?? null,
      table: d.table ?? [],
    };
  }
}
