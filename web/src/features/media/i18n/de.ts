import type { en } from './en';

/** German UI texts of the media feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.media.title': 'Medienserver',
  'widget.media.description': 'Was gerade läuft und was neu ist auf Jellyfin oder Plex.',
  'settings.section.media': 'Medienserver',
  'media.connect_under_settings': 'Verbinde Jellyfin oder Plex unter Einstellungen → Medienserver.',
  'media.now_playing': 'Läuft gerade',
  'media.nothing_playing': 'Gerade läuft nichts.',
  'media.recently_added': 'Neu hinzugefügt',
  'media.server': 'Medienserver',
  'media.url': '{name}-Adresse',
  'media.jellyfin_key': 'Zugangsschlüssel {value}',
  'media.plex_token': 'Plex-Token {value}',
  'media.jellyfin_hint':
    'Lege einen API-Schlüssel in Jellyfin unter Dashboard → API-Schlüssel an. Er wird verschlüsselt gespeichert; Cover lädt Waylight selbst.',
  'media.plex_hint':
    'Nutze das X-Plex-Token deines Kontos (Plex-Hilfeartikel „Finding an authentication token“). Es wird verschlüsselt gespeichert; Cover lädt Waylight selbst.',
  'media.setup_token': 'Zugangsschlüssel oder Token',
  'media.setup_step_jellyfin':
    'Jellyfin: Öffne Dashboard → API-Schlüssel, klicke + und nenne den Schlüssel Waylight.',
  'media.setup_step_plex':
    'Plex: Öffne in Plex Web einen beliebigen Titel, wähle ⋯ → Info abrufen → XML anzeigen; das Token ist der Wert nach X-Plex-Token= in der Adresszeile.',
  'media.setup_step_copy': 'Kopiere Schlüssel oder Token und füge ihn unten ein.',
  'media.setup_found': '{recent} neu hinzugefügt, {playing} laufen gerade',
};
