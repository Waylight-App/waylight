import { resolveCurrent } from '../../features/shopping/domain/shopping';
import { t } from '../../shared/i18n';
import { defineSearch, terms } from '../searchKit';

/** Longest shopping list in an answer; a list is read in full more often than other lists. */
const MAX_SHOPPING = 10;

/** The open items of the shopping list: "einkaufsliste", "what do i need to buy". */
export const search = defineSearch({
  module: 'shopping',
  intents: [
    {
      id: 'shopping.list',
      help: {
        group: 'day',
        de: 'Was auf deiner Einkaufsliste steht',
        en: 'What is on your shopping list',
        examples: { de: ['einkaufsliste', 'was muss ich einkaufen'], en: ['shopping list', 'groceries'] },
      },
      source: 'shopping',
      prefixes: ['einkauf', 'shopping'],
      answer: 'list',
      widgetKind: 'shopping',
      terms: [
        ...terms(3, 'einkaufsliste', 'einkaufen', 'einkauf', 'shopping', 'groceries', 'grocery'),
        ...terms(2, 'buy', 'besorgen', 'supermarkt', 'supermarket'),
        ...terms(1, 'liste', 'list', 'fehlt', 'missing'),
      ],
    },
  ],
  source: (c) => ({
    id: 'shopping',
    cacheMs: 30_000,
    async resolve(m) {
      // The list the widget shows: the saved choice, else the first one.
      const current = resolveCurrent(await c.shopping.overview());
      const items = current ? (await c.shopping.source(current).items()).filter((i) => !i.done) : [];
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'shopping',
        title: t('smart.shopping', { n: items.length }),
        at: null,
        empty: t('smart.shopping_empty'),
        items: items.slice(0, MAX_SHOPPING).map((i) => ({ label: i.summary, value: '' })),
      };
    },
  }),
});
