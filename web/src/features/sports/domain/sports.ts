/** A club (#285). */
export interface SportsTeam {
  id: number;
  name: string;
  shortName: string;
}

export interface SportsMatch {
  kickoff: Date;
  home: SportsTeam;
  away: SportsTeam;
  finished: boolean;
  hasScore: boolean;
  homeGoals: number;
  awayGoals: number;
  matchday: string;
}

export interface Standing {
  position: number;
  team: SportsTeam;
  played: number;
  points: number;
  goalDiff: number;
}

export interface League {
  id: string;
  name: string;
}

export interface SportsOverview {
  league: League;
  season: number;
  team: SportsTeam | null;
  last: SportsMatch | null;
  next: SportsMatch | null;
  position: Standing | null;
  table: Standing[];
}

/** Widget configuration: league and optional team (0 = league view). */
export function parseConfig(raw: Record<string, unknown>): { league: string; team: number } {
  const league =
    typeof raw['league'] === 'string' && /^[a-z0-9]{2,10}$/.test(raw['league']) ? raw['league'] : 'bl1';
  const team =
    typeof raw['team'] === 'number' && Number.isInteger(raw['team']) && raw['team'] > 0 ? raw['team'] : 0;
  return { league, team };
}

/** "2:1" for a match with a score, "" otherwise. */
export const score = (m: SportsMatch): string => (m.hasScore ? `${m.homeGoals}:${m.awayGoals}` : '');

/** Outcome of a finished match for a team: "win", "draw" or "loss". */
export function outcome(m: SportsMatch, teamId: number): 'win' | 'draw' | 'loss' | '' {
  if (!m.finished || !m.hasScore) return '';
  const own = m.home.id === teamId ? m.homeGoals : m.awayGoals;
  const other = m.home.id === teamId ? m.awayGoals : m.homeGoals;
  return own > other ? 'win' : own < other ? 'loss' : 'draw';
}

/** "+9" / "-3" / "0". */
export const signed = (n: number): string => (n > 0 ? `+${n}` : String(n));
