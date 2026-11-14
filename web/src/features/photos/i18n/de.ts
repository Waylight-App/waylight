import type { en } from './en';

/** German UI texts of the photo frame; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.photo_frame.title': 'Bilderrahmen',
  'widget.photo_frame.description': 'Deine eigenen Bilder im Wechsel – hier hochgeladen, ohne Foto-Server.',
  'photos.empty': 'Noch keine Bilder in „{album}“. Füge welche in den Widget-Einstellungen hinzu.',
  'photos.picture_of': 'Bild {n} von {total}',
  'photos.album': 'Album',
  'photos.interval': 'Sekunden pro Bild',
  'photos.fit': 'Darstellung',
  'photos.fit_contain': 'Ganzes Bild zeigen',
  'photos.fit_cover': 'Widget füllen',
  'photos.pictures': 'Bilder in „{album}“',
  'photos.upload': 'Bilder hochladen …',
  'photos.uploading': 'Lade {done} von {total} hoch …',
  'photos.upload_failed': '{n} Bild(er) nicht hinzugefügt, z. B. „{name}“: {reason}',
  'photos.move_earlier': 'Bild {n} nach vorne',
  'photos.move_later': 'Bild {n} nach hinten',
  'photos.delete': 'Bild {n} löschen',
  'photos.confirm_delete': 'Dieses Bild löschen?',
  'photos.usage': '{used} von {quota} belegt · {count} von {max} Bildern (alle Alben)',
  'photos.usage_label': 'Belegter Fotospeicher',
  'photos.backup_note':
    'Bilder werden auf 1920 px verkleinert und ohne Ortsdaten gespeichert. Backups enthalten sie, Datenexporte nicht.',
};
