import type { en } from './en';

/** German UI texts of the fuel feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.fuel.title': 'Spritpreise',
  'widget.fuel.description':
    'Günstigste Tankstellen in der Nähe oder deine Lieblingstankstellen (Deutschland, Tankerkönig).',
  'settings.section.fuel': 'Spritpreise',
  'fuel.connect_under_settings': 'Spritpreise in deiner Nähe.',
  'fuel.closed': 'geschlossen',
  'fuel.favourite': '{name} merken',
  'fuel.no_favourites': 'Markiere Tankstellen mit ☆, um sie hier zu sehen.',
  'fuel.no_stations': 'Keine Tankstellen im Umkreis.',
  'fuel.attribution': 'Daten: MTS-K über Tankerkönig (CC BY 4.0)',
  'fuel.view': 'Anzeigen',
  'fuel.view_cheapest': 'Günstigste Tankstellen',
  'fuel.view_favourites': 'Lieblingstankstellen',
  'fuel.search_place': 'Ort suchen',
  'fuel.location': 'Ort: {place} ({lat}, {lon})',
  'fuel.choose_location': 'Bitte wähle einen Ort.',
  'fuel.lat': 'Breitengrad',
  'fuel.lon': 'Längengrad',
  'fuel.fuel': 'Kraftstoff',
  'fuel.radius': 'Umkreis in km',
  'fuel.api_key': 'Tankerkönig-Zugangsschlüssel {value}',
  'fuel.key_hint':
    'Einen kostenlosen API-Schlüssel gibt es auf creativecommons.tankerkoenig.de. Preise werden höchstens alle fünf Minuten geladen und nur, solange das Widget offen ist.',
  'fuel.invalid_radius': 'Der Umkreis muss zwischen 1 und 25 km liegen.',
  'fuel.setup_place': 'Ort',
  'fuel.setup_place_hint': 'Ein Ort oder eine Postleitzahl. Waylight sucht Tankstellen in der Nähe.',
  'fuel.setup_api_key': 'Zugangsschlüssel',
  'fuel.setup_step_register':
    'Beantrage bei Tankerkönig mit deiner E-Mail-Adresse einen kostenlosen API-Schlüssel.',
  'fuel.setup_step_mail': 'Bestätige den Link in der E-Mail; der Schlüssel kommt mit einer zweiten E-Mail.',
  'fuel.setup_step_copy': 'Kopiere den Schlüssel und füge ihn unten ein.',
  'fuel.setup_place_not_found':
    'Kein Ort mit diesem Namen gefunden. Versuche eine Stadt oder eine Postleitzahl.',
  'fuel.setup_found': '{n} Tankstellen im Umkreis von {km} km um {place}',
  'fuel.setup_cheapest': 'Am günstigsten {fuel}: {name}, {price}',
};
