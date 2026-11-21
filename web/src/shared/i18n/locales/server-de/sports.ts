import type { ServerMessageKey } from '../../../api-client/generated/server-messages/sports';

/** German translations of the backend messages of `sports` (#285). */
export const serverDe: Record<ServerMessageKey, string> = {
  // Choose one of the offered leagues.
  srv_acb1d2af2955f9e7: 'Wähle eine der angebotenen Ligen.',
  // The team does not play in this league this season.
  srv_daddecf23a5f7008: 'Das Team spielt in dieser Saison nicht in dieser Liga.',
  // Open the widget settings and choose a league and a team.
  srv_46802779e0dc7472: 'Öffne die Widget-Einstellungen und wähle eine Liga und ein Team.',
  // The last result, next match and table position of your team, or the top of the table.
  srv_fbbaa2f9b927ab85:
    'Das letzte Ergebnis, das nächste Spiel und der Tabellenplatz deines Teams oder die Tabellenspitze.',
};
