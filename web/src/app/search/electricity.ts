import {
  currentHour,
  formatCent,
  parseConfig,
  zoneName,
} from '../../features/electricity/domain/electricity';
import { intlLocale, t } from '../../shared/i18n';
import { defineSearch, terms } from '../searchKit';

/** Dynamic electricity price: "strompreis", "wann ist strom am günstigsten". */
export const search = defineSearch({
  module: 'electricity',
  intents: [
    {
      id: 'electricity.price',
      help: {
        group: 'day',
        de: 'Was Strom gerade kostet und wann er am günstigsten ist',
        en: 'What electricity costs now and when it is cheapest',
        examples: {
          de: ['strompreis', 'wann ist strom günstig'],
          en: ['electricity price', 'cheapest power'],
        },
      },
      source: 'electricity',
      prefixes: ['strom', 'power'],
      answer: 'metric',
      widgetKind: 'electricity',
      terms: [
        ...terms(3, 'strom', 'strompreis', 'strompreise', 'electricity', 'boersenstrom', 'kwh'),
        ...terms(2, 'power'),
        ...terms(1.5, 'preis', 'price', 'kostet', 'costs', 'cost', 'gunstig', 'guenstig', 'billig', 'cheap'),
        ...terms(1.5, 'gunstigsten', 'guenstigsten', 'billigsten', 'cheapest'),
      ],
    },
  ],
  source: (c, ctx) => ({
    id: 'electricity',
    cacheMs: 300_000,
    async resolve(m) {
      const { zone } = parseConfig(ctx.widgetConfig('electricity') ?? {});
      const p = await c.electricity.prices(zone);
      const now = new Date();
      const locale = intlLocale();
      const hour = currentHour(p.hours, now);
      const ahead = p.hours.filter((h) => h.start.getTime() + 3_600_000 > now.getTime());
      const cheapest = ahead.reduce<(typeof ahead)[number] | null>(
        (best, h) => (!best || h.centPerKWh < best.centPerKWh ? h : best),
        null,
      );
      const time = (d: Date) => d.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' });
      return {
        kind: 'metric',
        intent: m.intent.id,
        widgetKind: 'electricity',
        title: t('smart.electricity', { zone: zoneName(zone) }),
        value: hour ? formatCent(hour.centPerKWh, locale) : t('smart.no_prices'),
        detail: cheapest
          ? t('smart.electricity_cheapest', {
              time: time(cheapest.start),
              price: formatCent(cheapest.centPerKWh, locale),
            })
          : undefined,
        at: p.fetchedAt,
      };
    },
  }),
});
