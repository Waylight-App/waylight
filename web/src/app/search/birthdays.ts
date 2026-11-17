import { t } from '../../shared/i18n';
import { defineSearch, inDays, MAX_ITEMS, terms } from '../searchKit';

/** Upcoming birthdays: "wer hat geburtstag", "birthdays". */
export const search = defineSearch({
  module: 'birthdays',
  intents: [
    {
      id: 'birthdays.upcoming',
      help: {
        group: 'day',
        de: 'Wer als Nächstes Geburtstag hat',
        en: 'Whose birthday is next',
        examples: { de: ['wer hat geburtstag', 'geburtstage'], en: ['birthdays', 'whose birthday'] },
      },
      source: 'birthdays',
      prefixes: ['geburtstag', 'birthday'],
      answer: 'list',
      widgetKind: 'birthdays',
      terms: [
        ...terms(3, 'geburtstag', 'geburtstage', 'birthday', 'birthdays', 'bday'),
        ...terms(1, 'whose', 'naechste', 'nachste', 'next', 'gratulieren'),
      ],
    },
  ],
  source: (c) => ({
    id: 'birthdays',
    cacheMs: 600_000,
    async resolve(m) {
      const list = await c.birthdays.upcoming(MAX_ITEMS);
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'birthdays',
        title: t('smart.birthdays'),
        at: null,
        empty: t('smart.no_birthdays'),
        items: list.map((b) => ({
          label: b.name,
          value: [inDays(b.inDays), b.age > 0 ? t('smart.turns', { age: b.age }) : '']
            .filter(Boolean)
            .join(' · '),
          state: b.inDays === 0 ? 'ok' : undefined,
        })),
      };
    },
  }),
});
