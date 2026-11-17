import { intlLocale, t } from '../../shared/i18n';
import { defineSearch, MAX_ITEMS, terms } from '../searchKit';

/** Fuel prices nearby (#53): "spritpreise", "wo ist diesel günstig". */
export const search = defineSearch({
  module: 'fuel',
  intents: [
    {
      id: 'fuel.prices',
      help: {
        group: 'day',
        de: 'Wo Sprit gerade am günstigsten ist',
        en: 'Where fuel is cheapest right now',
        examples: { de: ['spritpreise', 'diesel günstig'], en: ['fuel prices', 'cheap petrol'] },
      },
      source: 'fuel',
      prefixes: ['sprit', 'tanken', 'fuel'],
      answer: 'list',
      widgetKind: 'fuel',
      terms: [
        ...terms(3, 'sprit', 'spritpreis', 'spritpreise', 'benzin', 'diesel', 'e10', 'tanken'),
        ...terms(1, 'super'),
        ...terms(3, 'tankstelle', 'tankstellen', 'fuel', 'petrol', 'gasoline'),
        ...terms(
          1.5,
          'preis',
          'preise',
          'price',
          'prices',
          'gunstig',
          'guenstig',
          'billig',
          'cheap',
          'cheapest',
        ),
      ],
    },
  ],
  source: (c) => ({
    id: 'fuel',
    cacheMs: 300_000,
    async resolve(m) {
      const o = await c.fuel.overview();
      const locale = intlLocale();
      const price = (p: number | null) =>
        p === null ? t('smart.fuel_no_price') : `${p.toLocaleString(locale, { minimumFractionDigits: 3 })} €`;
      return {
        kind: 'list',
        intent: m.intent.id,
        widgetKind: 'fuel',
        title: t('smart.fuel', { fuel: o.fuel.toUpperCase(), place: o.place }),
        at: null,
        empty: t('smart.no_fuel'),
        items: (o.favourites.length > 0 ? o.favourites : o.cheapest).slice(0, MAX_ITEMS).map((s) => ({
          label: s.brand || s.name,
          value: [
            price(s.price),
            `${s.distanceKm.toLocaleString(locale, { maximumFractionDigits: 1 })} km`,
          ].join(' · '),
          state: s.open ? undefined : 'off',
        })),
      };
    },
  }),
});
