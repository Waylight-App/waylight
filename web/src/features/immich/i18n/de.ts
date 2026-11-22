import type { en } from './en';

/** German UI texts of the immich feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.immich.title': 'Immich',
  'widget.immich.description': 'Statistik der Fotobibliothek und Erinnerungen des Tages.',
  'settings.section.immich': 'Immich',
  'immich.connect_under_settings': 'Verbinde Immich unter Einstellungen → Immich.',
  'immich.photos': 'Fotos',
  'immich.videos': 'Videos',
  'immich.storage': 'Speicher',
  'immich.years_ago': 'Heute vor {n} Jahren',
  'immich.memory_photo': 'Foto von vor {n} Jahren',
  'immich.show_memories': 'Erinnerungen des Tages anzeigen',
  'immich.url': 'Immich-Adresse',
  'immich.api_key': 'Zugangsschlüssel {value}',
  'immich.key_hint':
    'Lege einen API-Schlüssel unter Kontoeinstellungen → API-Schlüssel mit den Berechtigungen asset.statistics, server.storage, memory.read und asset.view an. Er wird verschlüsselt gespeichert; Fotos lädt Waylight selbst.',
  'immich.setup_step_profile': 'Öffne Immich und klicke oben rechts auf dein Profilbild.',
  'immich.setup_step_keys': 'Wähle Kontoeinstellungen → API-Schlüssel → Neuer API-Schlüssel.',
  'immich.setup_step_permissions':
    'Nenne ihn Waylight, hake asset.statistics, server.storage, memory.read und asset.view an und klicke Erstellen.',
  'immich.setup_step_copy': 'Kopiere den Schlüssel und füge ihn unten ein.',
  'immich.setup_key': 'Zugangsschlüssel',
  'immich.setup_photos': '{n} Fotos, {videos} Videos',
};
