import { describe, expect, it } from 'vitest';
import { outcome, parseConfig, score, signed, type SportsMatch } from './sports';

const m = (homeGoals: number, awayGoals: number, finished = true): SportsMatch => ({
  kickoff: new Date(),
  home: { id: 7, name: 'Borussia Dortmund', shortName: 'BVB' },
  away: { id: 40, name: 'FC Bayern München', shortName: 'Bayern' },
  finished,
  hasScore: true,
  homeGoals,
  awayGoals,
  matchday: '10. Spieltag',
});

describe('sports (#285)', () => {
  it('reads the widget configuration', () => {
    expect(parseConfig({})).toEqual({ league: 'bl1', team: 0 });
    expect(parseConfig({ league: 'bl2', team: 40 })).toEqual({ league: 'bl2', team: 40 });
    expect(parseConfig({ league: '../x', team: -1 })).toEqual({ league: 'bl1', team: 0 });
  });
  it('describes results from the team’s point of view', () => {
    expect(score(m(1, 2))).toBe('1:2');
    expect(outcome(m(1, 2), 40)).toBe('win');
    expect(outcome(m(1, 2), 7)).toBe('loss');
    expect(outcome(m(1, 1), 7)).toBe('draw');
    expect(outcome(m(0, 0, false), 7)).toBe('');
    expect(signed(9)).toBe('+9');
    expect(signed(-3)).toBe('-3');
  });
});
