import type { en } from './en';

/** German UI texts of the documents feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.documents.title': 'Garantie und Dokumente',
  'widget.documents.description':
    'Belege und Garantieunterlagen mit dem Ende jeder Garantie; Fotos und PDFs bleiben auf deinem Server.',
  'compact.example.documents': 'Kopfhörer: Garantie endet am 12. Dez.',
  'documents.new': 'Neues Dokument',
  'documents.edit': 'Dokument ändern',
  'documents.title': 'Produkt',
  'documents.purchased_on': 'Gekauft am',
  'documents.warranty_end': 'Garantie endet am',
  'documents.note': 'Notiz',
  'documents.file': 'Foto oder PDF (bis 10 MB)',
  'documents.file_replace': 'Wähle eine neue Datei, um {name} zu ersetzen.',
  'documents.status.running': 'Garantie bis {date}',
  'documents.status.ending': 'Garantie endet am {date}',
  'documents.status.ended': 'Garantie endete am {date}',
  'documents.status.none': 'Keine Garantie eingetragen',
  'documents.empty': 'Noch keine Dokumente.',
  'documents.download': '{name} herunterladen',
  'documents.edit_named': '{name} ändern',
  'documents.remove': '{name} entfernen',
  'documents.error.title': 'Bitte einen Produktnamen mit bis zu 80 Zeichen eingeben.',
  'documents.error.dates': 'Die Garantie kann nicht vor dem Kauf enden.',
  'documents.error.file_type': 'Bitte ein Foto (JPEG, PNG oder WebP) oder ein PDF wählen.',
  'documents.error.file_size': 'Die Datei darf nicht leer oder größer als 10 MB sein.',
};
