/** English UI texts of the sports widget (#285). */
export const en = {
  'widget.sports.title': 'Sports',
  'widget.sports.description': 'Last result, next fixture and table position of your team.',
  'sports.league': 'League',
  'sports.team': 'Team',
  'sports.no_team': 'No team: show the top of the table',
  'sports.last': 'Last result',
  'sports.next': 'Next match',
  'sports.position': '{position}. place · {points} points · {played} matches',
  'sports.none': 'No match',
  'sports.points': 'Pts',
  'sports.goal_diff': 'Diff',
  'sports.win': 'won',
  'sports.draw': 'draw',
  'sports.loss': 'lost',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
