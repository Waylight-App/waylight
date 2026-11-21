/** English UI texts of the family board feature (#685). */
export const en = {
  'widget.board.title': 'Family board',
  'widget.board.description': 'Short notes for everyone, pinned or with a last day.',
  'compact.example.board': '3 notes',
  'board.new': 'New note',
  'board.edit': 'Change note',
  'board.text': 'Note',
  'board.expires': 'Show until',
  'board.pin': 'Pin to the top',
  'board.unpin': 'Unpin',
  'board.pinned': 'Pinned',
  'board.until': 'until {date}',
  'board.remove': 'Remove note',
  'board.confirm_remove': 'Remove this note?',
  'board.empty': 'Nothing on the board yet.',
  'board.error.text': 'Write a note of up to 500 characters.',
  'board.error.expiry': 'Choose today or a later day.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
