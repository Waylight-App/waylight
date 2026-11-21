import type { ServerMessageKey } from '../../../api-client/generated/server-messages/weather';

/**
 * German translations of the backend error messages of `weather` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // Latitude and longitude are required.
  srv_9d350adc128ef832: 'Breiten- und Längengrad sind erforderlich.',
  // Please enter between 2 and 100 characters.
  srv_990b607e23145ec7: 'Bitte zwischen 2 und 100 Zeichen eingeben.',
  // The latitude must be between -90 and 90.
  srv_3f60e80ed2350494: 'Der Breitengrad muss zwischen -90 und 90 liegen.',
  // The longitude must be between -180 and 180.
  srv_39566e8cbb1ec35d: 'Der Längengrad muss zwischen -180 und 180 liegen.',
  // The rain radar is only available for Germany.
  srv_d5d2cea1e1d1fa8d: 'Das Regenradar ist nur für Deutschland verfügbar.',
  // The weather warnings are not available right now.
  srv_314a73e152936a17: 'Die Wetterwarnungen sind gerade nicht verfügbar.',
  // The pollen and UV forecasts are not available right now.
  srv_cb30f7793468eebc: 'Pollenflug und UV-Index sind gerade nicht verfügbar.',
  // Clear sky
  srv_079ecf8df23a0daf: 'Klarer Himmel',
  // Dense drizzle
  srv_0d9429b1a019413e: 'Starker Nieselregen',
  // Drizzle
  srv_aa17e4cc0cbb8ecd: 'Nieselregen',
  // Fog
  srv_f2bb8d199c92d43f: 'Nebel',
  // Freezing drizzle
  srv_061bd69a46d49c8f: 'Gefrierender Nieselregen',
  // Freezing rain
  srv_d8b57a87d52d6ded: 'Gefrierender Regen',
  // Heavy rain
  srv_385fbf7e52385aae: 'Starker Regen',
  // Heavy snow
  srv_68c0138424f8074b: 'Starker Schneefall',
  // Light drizzle
  srv_6793a3117aad7cf5: 'Leichter Nieselregen',
  // Light rain
  srv_9f8d7b03da2db18b: 'Leichter Regen',
  // Light snow
  srv_fd8baefca48339ea: 'Leichter Schneefall',
  // Mainly clear
  srv_5ae1ab5870ccaffa: 'Überwiegend klar',
  // Overcast
  srv_74a09b179be18310: 'Bedeckt',
  // Partly cloudy
  srv_c72683131dff0339: 'Teilweise bewölkt',
  // Rain
  srv_984afe2bd363c533: 'Regen',
  // Rain showers
  srv_c09988829a692b1c: 'Regenschauer',
  // Rime fog
  srv_a5e43404e0b672b6: 'Raureifnebel',
  // Snow
  srv_8300b224ee074d12: 'Schneefall',
  // Snow grains
  srv_852f4060b58e9e0e: 'Schneegriesel',
  // Snow showers
  srv_5184b87cd74a5e31: 'Schneeschauer',
  // Thunderstorm
  srv_11111027a2159b52: 'Gewitter',
  // Thunderstorm with hail
  srv_95215781350e5dbc: 'Gewitter mit Hagel',
  // Unknown
  srv_c592307ea80f16b9: 'Unbekannt',
  // Violent showers
  srv_60e5e8ca070c80f3: 'Heftige Schauer',
  // {temp}°
  srv_d72ab051081647bd: '{temp}°',
  // A rain radar map of the last and next hours around a place in Germany.
  srv_194aeb9844f81b54:
    'Eine Regenradar-Karte der letzten und nächsten Stunden rund um einen Ort in Deutschland.',
  // Open the widget settings and pick a saved place or search for one. No account or key is needed.
  srv_25ba727a1f2b2d3d:
    'Öffne die Widget-Einstellungen und wähle einen gespeicherten Ort oder suche einen. Ein Konto oder Schlüssel ist nicht nötig.',
  // Open the widget settings and pick the place the map should centre on.
  srv_7b93225b8681a679:
    'Öffne die Widget-Einstellungen und wähle den Ort, um den die Karte zentriert sein soll.',
  // The forecast for a place: today, the next hours and days, and in Germany official weather warnings.
  srv_fef0c29ef6184c8d:
    'Die Vorhersage für einen Ort: heute, die nächsten Stunden und Tage und in Deutschland die amtlichen Wetterwarnungen.',
};
