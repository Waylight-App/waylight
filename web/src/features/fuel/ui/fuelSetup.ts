import { intlLocale, t } from '../../../shared/i18n';
import type { SetupDescriptor } from '../../../shared/setup/setup';
import { fuelName, type PlaceResult } from '../domain/fuel';
import type { FuelUseCases } from '../usecase/fuelUseCases';

/** A pump price with its three decimals, e.g. 1,739 €. */
const euros = (price: number | null) =>
  (price ?? 0).toLocaleString(intlLocale(), { style: 'currency', currency: 'EUR', minimumFractionDigits: 3 });

/** Defaults of the settings page; it stays the place to change them. */
const DEFAULT_RADIUS_KM = 5;

/**
 * The setup assistant of the fuel price widget (#618), as data. Tankerkönig
 * is a public service, so the first step asks for the place instead of an
 * address; the place search of the weather module turns it into coordinates.
 */
export function fuelSetup(
  fuel: FuelUseCases,
  searchPlaces: (query: string) => Promise<PlaceResult[]>,
): SetupDescriptor {
  return {
    service: 'Tankerkönig',
    fields: [
      {
        key: 'place',
        hint: 'fuel.setup_place_hint',
        label: 'fuel.setup_place',
        stage: 'address',
        type: 'text',
        placeholder: 'Berlin',
      },
      {
        key: 'apiKey',
        label: 'fuel.setup_api_key',
        stage: 'credential',
        type: 'password',
        placeholder: '00000000-0000-0000-0000-000000000000',
      },
    ],
    steps: ['fuel.setup_step_register', 'fuel.setup_step_mail', 'fuel.setup_step_copy'],
    credentialPage: () => 'https://onboarding.tankerkoenig.de/',
    async check(values) {
      const [place] = await searchPlaces(values['place'] ?? '');
      if (!place) throw new Error(t('fuel.setup_place_not_found'));
      await fuel.saveSettings({
        apiKey: values['apiKey'] ?? '',
        lat: place.lat,
        lon: place.lon,
        place: place.name,
        radiusKm: DEFAULT_RADIUS_KM,
        fuel: 'e10',
      });
      const o = await fuel.overview();
      const cheapest = o.cheapest.find((s) => s.price !== null);
      return {
        lines: [
          t('fuel.setup_found', { n: o.cheapest.length, km: DEFAULT_RADIUS_KM, place: place.name }),
          ...(cheapest
            ? [
                t('fuel.setup_cheapest', {
                  fuel: fuelName(o.fuel),
                  name: cheapest.name,
                  price: euros(cheapest.price),
                }),
              ]
            : []),
        ],
      };
    },
  };
}
