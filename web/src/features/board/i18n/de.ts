import type { en } from './en';

/** German UI texts of the family board feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'widget.board.title': 'Familien-Pinnwand',
  'widget.board.description': 'Kurze Notizen für alle, angeheftet oder mit letztem Tag.',
  'compact.example.board': '3 Notizen',
  'board.new': 'Neue Notiz',
  'board.edit': 'Notiz ändern',
  'board.text': 'Notiz',
  'board.expires': 'Anzeigen bis',
  'board.pin': 'Oben anheften',
  'board.unpin': 'Lösen',
  'board.pinned': 'Angeheftet',
  'board.until': 'bis {date}',
  'board.remove': 'Notiz entfernen',
  'board.confirm_remove': 'Diese Notiz entfernen?',
  'board.empty': 'Noch nichts an der Pinnwand.',
  'board.error.text': 'Bitte eine Notiz mit bis zu 500 Zeichen schreiben.',
  'board.error.expiry': 'Bitte heute oder einen späteren Tag wählen.',
};
