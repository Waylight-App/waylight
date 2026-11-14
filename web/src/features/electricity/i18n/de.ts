import type { en } from './en';

/** German UI texts of the electricity price widget; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.electricity.title': 'Strompreise',
  'widget.electricity.description':
    'Börsenstrompreis pro Stunde für heute und morgen, mit den günstigsten Stunden.',
  'electricity.zone': 'Gebotszone',
  'electricity.zone_de': 'Deutschland und Luxemburg',
  'electricity.zone_at': 'Österreich',
  'electricity.now': 'Jetzt',
  'electricity.today': 'Heute',
  'electricity.tomorrow': 'Morgen',
  'electricity.cheapest': 'Günstigste Stunden',
  'electricity.stale': 'Die Quelle antwortet nicht; das sind die zuletzt bekannten Preise.',
  'electricity.market_note': 'Börsenpreis ohne Steuern und Abgaben.',
  'electricity.hour': '{time}: {price}',
  'electricity.tomorrow_later': 'Die Preise für morgen folgen gegen 13 Uhr.',
};
