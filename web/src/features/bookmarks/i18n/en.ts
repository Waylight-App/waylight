/** English UI texts added to the bookmarks feature (#120, #193). */
export const en = {
  'bookmarks.sort': 'Sort order',
  'bookmarks.sort_manual': 'Default',
  'bookmarks.sort_az': 'A–Z',
  'bookmarks.sort_newest': 'Newest first',
  'widget.quicklinks.title': 'Quick links',
  'widget.quicklinks.description': 'Your favourite bookmarks as tiles; 1 to 9 open them',
  'quicklinks.empty': 'No bookmarks yet. Add them in a bookmarks widget or with "b" and a link in the search.',
  'quicklinks.key_hint': '{title} (key {key})',
  'quicklinks.settings_hint':
    'Choose up to {max} bookmarks. The first {keys} open with the keys 1 to {keys}. Without a choice the widget shows your first bookmarks.',
  'quicklinks.order': 'Order of the quick links',
  'quicklinks.move_up': 'Move {title} up',
  'quicklinks.move_down': 'Move {title} down',
  'quicklinks.choose': 'Bookmarks ({n} of {max})',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
