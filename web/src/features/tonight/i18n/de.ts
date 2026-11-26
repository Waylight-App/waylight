import type { en } from './en';

/** German UI texts of the Tonight card; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'tonight.label': 'Uhrzeit und Wetter',
  'tonight.choose_a_location': 'Ort in den Widget-Einstellungen wählen',
  'tonight.weather_unavailable': 'Wetter nicht verfügbar',
  'tonight.loading': 'Wetter wird geladen',
  'tonight.more_warnings': '+{n} weitere',
  'tonight.warning': 'Wetterwarnung',
  'tonight.title_in': 'Heute Abend in {place}',
  'tonight.title': 'Heute Abend',
  'tonight.feels': 'gefühlt {t}',
  'tonight.forecast': 'Vorhersage',
};
