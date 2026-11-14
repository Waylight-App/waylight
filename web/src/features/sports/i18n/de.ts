import type { en } from './en';

/** German UI texts of the sports widget; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.sports.title': 'Sport',
  'widget.sports.description': 'Letztes Ergebnis, nächstes Spiel und Tabellenplatz deines Teams.',
  'sports.league': 'Liga',
  'sports.team': 'Team',
  'sports.no_team': 'Kein Team: Tabellenspitze zeigen',
  'sports.last': 'Letztes Ergebnis',
  'sports.next': 'Nächstes Spiel',
  'sports.position': '{position}. Platz · {points} Punkte · {played} Spiele',
  'sports.none': 'Kein Spiel',
  'sports.points': 'Pkt.',
  'sports.goal_diff': 'Diff.',
  'sports.win': 'gewonnen',
  'sports.draw': 'unentschieden',
  'sports.loss': 'verloren',
};
