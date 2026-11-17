import { parseConfig, score } from '../../features/sports/domain/sports';
import { intlLocale, t } from '../../shared/i18n';
import { defineSearch, terms } from '../searchKit';

/** Football (#56): "spielt werder heute", "bundesliga tabelle", "next match". */
export const search = defineSearch({
  module: 'sports',
  intents: [
    {
      id: 'sports.team',
      help: {
        group: 'day',
        de: 'Letztes und nächstes Spiel deines Teams und der Tabellenplatz',
        en: 'Last and next match of your team and its place in the table',
        examples: {
          de: ['wann spielt werder', 'bundesliga tabelle'],
          en: ['next match', 'football results'],
        },
      },
      source: 'sports',
      prefixes: ['fussball', 'football'],
      answer: 'list',
      widgetKind: 'sports',
      terms: [
        ...terms(3, 'fussball', 'football', 'soccer', 'bundesliga', 'liga', 'league', 'tabelle', 'table'),
        ...terms(3, 'spiel', 'spiele', 'spielt', 'gespielt', 'match', 'matches', 'anpfiff', 'kickoff'),
        ...terms(2, 'ergebnis', 'ergebnisse', 'results', 'score', 'tor', 'tore', 'goals', 'gewonnen', 'won'),
        ...terms(1, 'naechstes', 'nachstes', 'next', 'letztes', 'last'),
      ],
    },
  ],
  source: (c, ctx) => ({
    id: 'sports',
    cacheMs: 300_000,
    async resolve(m) {
      const cfg = parseConfig(ctx.widgetConfig('sports') ?? {});
      const o = await c.sports.overview(cfg.league, cfg.team);
      const when = (d: Date) =>
        d.toLocaleString(intlLocale(), {
          weekday: 'short',
          day: 'numeric',
          month: 'short',
          hour: '2-digit',
          minute: '2-digit',
        });
      const items: { label: string; value: string }[] = [];
      if (o.last)
        items.push({ label: `${o.last.home.shortName} – ${o.last.away.shortName}`, value: score(o.last) });
      if (o.next)
        items.push({
          label: `${o.next.home.shortName} – ${o.next.away.shortName}`,
          value: when(o.next.kickoff),
        });
      if (o.position)
        items.push({
          label: t('smart.sports_position'),
          value: `${o.position.position}. · ${o.position.points} P`,
        });
      if (!o.team)
        for (const s of o.table.slice(0, 5))
          items.push({ label: `${s.position}. ${s.team.shortName}`, value: `${s.points} P` });
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'sports',
        title: o.team ? o.team.name : o.league.name,
        at: null,
        empty: t('smart.no_matches'),
        items,
      };
    },
  }),
});
