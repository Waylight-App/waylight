import type { en } from './en';

/** German UI texts of waylight.yaml; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'transfer.config_title': 'Einstellungen als YAML',
  'transfer.config_intro':
    'Aussehen, Suche, deine Dienst-Kacheln, Lesezeichen und das Dashboard als eine lesbare Datei, für Git oder eine zweite Instanz. Ohne Geheimnisse; ${NAME} wird aus der Server-Umgebung gelesen.',
  'transfer.config_download': 'waylight.yaml herunterladen',
  'transfer.config_apply': 'waylight.yaml anwenden…',
  'transfer.config_confirm':
    'Jeder Abschnitt der Datei ersetzt diese Einstellungen vollständig; fehlende Abschnitte bleiben, wie sie sind. Fortfahren?',
  'transfer.config_managed':
    'Von CONFIG_FILE verwaltet und hier schreibgeschützt: {sections}. Ändere sie in der Datei und starte Waylight neu.',
  'transfer.foreign_title': 'Aus einem anderen Dashboard importieren',
  'transfer.foreign_intro':
    'Fügt die Dienst-Kacheln, Lesezeichen und unterstützten Widgets aus Homepage oder Homarr hinzu. Vorhandenes bleibt unverändert, ein zweiter Import schadet also nicht. Du siehst zuerst eine Vorschau.',
  'transfer.foreign_format': 'Dashboard',
  'transfer.foreign_homepage': 'Homepage (gethomepage.dev)',
  'transfer.foreign_homarr': 'Homarr 0.x',
  'transfer.foreign_choose': 'Dateien wählen…',
  'transfer.foreign_hint_homepage':
    'Wähle services.yaml, bookmarks.yaml und widgets.yaml aus dem Konfigurationsordner von Homepage.',
  'transfer.foreign_hint_homarr':
    'Wähle die Board-Datei aus dem Ordner data/configs von Homarr. Exporte von Homarr 1.x lassen sich nicht lesen.',
  'transfer.foreign_summary':
    'Fügt {tiles} Kacheln, {bookmarks} Lesezeichen und {widgets} Widgets hinzu; {skipped} gibt es schon.',
  'transfer.foreign_tab': 'Neuer Tab: {tab}.',
  'transfer.foreign_warnings': '{n} Dinge werden nicht importiert',
  'transfer.foreign_yaml': 'Ergebnis als waylight.yaml zeigen',
  'transfer.foreign_apply': 'Importieren',
};
