import { parseHolidaysConfig } from '../../features/holidays/domain/holidays';
import { intlLocale, t } from '../../shared/i18n';
import { defineSearch, inDays, MAX_ITEMS, terms } from '../searchKit';

/** Public and school holidays: "nächster feiertag", "schulferien". */
export const search = defineSearch({
  module: 'holidays',
  intents: [
    {
      id: 'holidays.next',
      help: {
        group: 'day',
        de: 'Die nächsten Feiertage und Ferien',
        en: 'The next public and school holidays',
        examples: { de: ['nächster feiertag', 'schulferien'], en: ['next holiday', 'school holidays'] },
      },
      source: 'holidays',
      prefixes: ['feiertag', 'ferien', 'holiday'],
      answer: 'list',
      widgetKind: 'holidays',
      terms: [
        ...terms(3, 'feiertag', 'feiertage', 'ferien', 'schulferien', 'holiday', 'holidays', 'brueckentag'),
        ...terms(1, 'frei', 'naechster', 'nachster', 'nachste', 'next', 'public', 'school'),
      ],
    },
  ],
  source: (c, ctx) => ({
    id: 'holidays',
    cacheMs: 3_600_000,
    async resolve(m) {
      const cfg = parseHolidaysConfig(ctx.widgetConfig('holidays') ?? {});
      const o = await c.holidays.upcoming(cfg.state, MAX_ITEMS);
      const date = (d: Date) => d.toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short' });
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'holidays',
        title: t('smart.holidays'),
        at: null,
        empty: o.schoolError || t('smart.no_holidays'),
        items: o.entries.slice(0, MAX_ITEMS).map((h) => ({
          label: h.name,
          value: h.current ? t('smart.holiday_now') : `${inDays(h.daysUntil)} · ${date(h.start)}`,
          state: h.current ? 'ok' : undefined,
        })),
      };
    },
  }),
});
