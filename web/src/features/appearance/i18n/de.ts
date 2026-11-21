import type { en } from './en';

/** German UI texts of the background picker; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'appearance.background_image': 'Hintergrund',
  'appearance.background_choose': 'Auswählen…',
  'appearance.background_title': 'Hintergrund',
  'appearance.background_scope': 'Wo der Hintergrund gilt',
  'appearance.background_scope_global': 'Alle Tabs',
  'appearance.background_scope_tab': 'Nur „{name}“',
  'appearance.background_none_tab': 'Nutzt den Hintergrund aller Tabs',
  'appearance.background_default': 'Standard-Hintergrund',
  'appearance.background_overridden':
    '„{name}“ hat einen eigenen Hintergrund, der stattdessen angezeigt wird.',
  'appearance.background_clear_tab': 'Hintergrund von „{name}“ entfernen, damit dieser auch dort erscheint',
  'appearance.background_builtin': 'Eingebaut',
  'appearance.background_builtin_dusk': 'Dämmerung',
  'appearance.background_builtin_dawn': 'Morgenrot',
  'appearance.background_builtin_forest': 'Wald',
  'appearance.background_builtin_ocean': 'Ozean',
  'appearance.background_builtin_aurora': 'Polarlicht',
  'appearance.background_builtin_sand': 'Sand',
  'appearance.background_builtin_slate': 'Schiefer',
  'appearance.background_builtin_paper': 'Papier',
  'appearance.background_uploads': 'Deine Bilder',
  'appearance.background_upload': 'Bild hochladen',
  'appearance.background_upload_pick': 'Dein Bild, {w} × {h}',
  'appearance.background_upload_hint':
    'JPEG, PNG oder GIF bis 25 MB. Wird auf diesem Server gespeichert, auf 2560 px verkleinert und ohne Ortsdaten abgelegt; Backups enthalten die Bilder.',
  'appearance.background_delete': 'Bild löschen',
  'appearance.background_delete_confirm':
    'Dieses Bild löschen? Wo es verwendet wird, gilt dann wieder der Standard-Hintergrund.',
  'appearance.background_url': 'Bild aus dem Web',
  'appearance.background_use_url': 'Verwenden',
  'appearance.background_reset': 'Standard-Hintergrund',
  'appearance.background_reset_tab': 'Hintergrund aller Tabs verwenden',
  'appearance.background_invalid': 'Gib die http(s)-Adresse eines Bildes ein.',
  'appearance.background_too_large': 'Das Bild ist größer als 25 MB.',
  'dashboard.background_button': 'Hintergrund',
  'appearance.accessibility': 'Barrierefreiheit',
  'appearance.text_size': 'Schriftgröße',
  'appearance.text_size_normal': 'Normal',
  'appearance.text_size_large': 'Groß',
  'appearance.text_size_larger': 'Größer',
  'appearance.high_contrast': 'Hoher Kontrast',
  'appearance.high_contrast_hint':
    'Schwarz und Weiß mit kräftigen Farben, für hell und dunkel. Ersetzt Theme, Akzentfarbe und Hintergrund, solange es an ist.',
  'appearance.simple_view': 'Einfache Ansicht',
  'appearance.simple_view_hint':
    'Eine Spalte mit dem kurzen Stand jedes Widgets, größere Schaltflächen und kein Bearbeitungsmodus. Zum Bearbeiten des Dashboards hier wieder ausschalten.',
};
