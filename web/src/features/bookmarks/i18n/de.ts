import type { en } from './en';

/** German UI texts added to the bookmarks feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'bookmarks.sort': 'Sortierung',
  'bookmarks.sort_manual': 'Standard',
  'bookmarks.sort_az': 'A–Z',
  'bookmarks.sort_newest': 'Neueste zuerst',
  'widget.quicklinks.title': 'Schnellzugriff',
  'widget.quicklinks.description': 'Deine liebsten Lesezeichen als Kacheln; 1 bis 9 öffnen sie',
  'quicklinks.empty':
    'Noch keine Lesezeichen. Lege sie in einem Lesezeichen-Widget an oder mit „b“ und einem Link in der Suche.',
  'quicklinks.key_hint': '{title} (Taste {key})',
  'quicklinks.settings_hint':
    'Wähle bis zu {max} Lesezeichen. Die ersten {keys} öffnen sich mit den Tasten 1 bis {keys}. Ohne Auswahl zeigt das Widget deine ersten Lesezeichen.',
  'quicklinks.order': 'Reihenfolge des Schnellzugriffs',
  'quicklinks.move_up': '{title} nach oben',
  'quicklinks.move_down': '{title} nach unten',
  'quicklinks.choose': 'Lesezeichen ({n} von {max})',
};
