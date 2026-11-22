import { locale, t } from '../shared/i18n';
import type { Container } from './container';
import type { Place } from '../features/onboarding/domain/profile';
import { exampleSetOf, searchExamples } from '../features/onboarding/domain/wizard';
import type { SmartSearch } from '../features/search/usecase/smartSearch';
import type { WeatherConfig } from '../features/weather/domain/weather';
import type { WizardContext } from '../features/onboarding/ui/wizard/wizardContext';

/** The default weather settings at a saved place; the weather answers need nothing more. */
const weatherAt = (p: Place): WeatherConfig => ({
  lat: p.lat,
  lon: p.lon,
  label: p.label,
  days: 3,
  warnings: true,
  pollenRegion: 0,
  pollenTypes: [],
  uvCity: '',
});

/** Used before the modules are known: nothing to try. */
export const noWizardSearch: WizardContext['search'] = {
  examples: async () => [],
  answer: async () => [],
};

/**
 * The search of the onboarding wizard's try-it step (#740): the app's own
 * smart search and help, so the examples answer like the real search does.
 * Loaded on first use; the wizard cannot act on anything (no actions).
 */
export function wizardSearch(c: Container, modules: Set<string>): WizardContext['search'] {
  let smart: SmartSearch | undefined;
  let place: Place | undefined;

  return {
    async examples(template) {
      return searchExamples(locale() === 'de' ? 'de' : 'en', exampleSetOf(template), modules);
    },
    async answer(query, signal, at) {
      place = at;
      if (!smart) {
        const { createSmartSearch } = await import('./searchSources');
        smart = createSmartSearch(c, modules, {
          canAct: false,
          weatherLocation: () => (place ? weatherAt(place) : null),
        });
      }
      // Before any widget is set up the sources fail; those cards say nothing useful here.
      const unavailable = t('smart.unavailable_title');
      return (await smart.answer(query, signal)).filter((a) => a.title !== unavailable);
    },
  };
}
