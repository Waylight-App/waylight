/** English texts of the wiki search (#321). */
export const en = {
  'settings.section.wiki': 'Wiki search',
  'app.group.wiki': 'Wiki',
  'wiki.intro':
    'Find wiki pages from the search field. Waylight asks the wiki while you type and stores no page content. Type “w ” first to search the wikis only.',
  'wiki.shared_token':
    'Everyone who uses this start page sees what the API token may read. Use a token of an account that only reads what all of you may see.',
  'wiki.none': 'No wiki connected yet.',
  'wiki.add': 'Connect a wiki',
  'wiki.name': 'Name',
  'wiki.kind': 'Wiki software',
  'wiki.url': 'Address of the wiki',
  'wiki.token': 'API token',
  'wiki.token_stored': 'stored – leave empty to keep it',
  'wiki.token_hint_bookstack':
    'In BookStack: your profile → API tokens. Enter it as “token ID:token secret”.',
  'wiki.token_hint_outline': 'In Outline: Settings → API & apps → New API key.',
  'wiki.token_hint_wikijs': 'In Wiki.js: Administration → API access → New API key (read access is enough).',
  'wiki.test': 'Test',
  'wiki.test_ok': 'The wiki answered.',
  'wiki.saved': 'Saved.',
  'wiki.delete': 'Remove',
  'wiki.failed': 'No answer from {names}.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
