/** English UI texts of the search feature that live with it (#477, #478). */
export const en = {
  'search.own_keywords': 'Own keywords',
  'search.own_keywords_hint':
    'Type a keyword alone in the search field and press Enter to open what it points to, for example “nas” for your NAS. Keywords work on every device you sign in on.',
  'search.own_keyword': 'Own keyword',
  'search.keyword_target_kind': 'Opens',
  'search.keyword_target': 'Target',
  'search.keyword_pick': 'Choose…',
  'search.keyword_query_placeholder': 'Search, e.g. waste next week',
  'search.kw_service': 'Service tile',
  'search.kw_bookmark': 'Bookmark',
  'search.kw_widget': 'Widget',
  'search.kw_url': 'Web address',
  'search.kw_query': 'Search',
  'search.delete_keyword': 'Delete keyword {keyword}',
  'search.keyword_result': 'Your keyword',
  'search.add_keyword': 'Add keyword…',
  'search.add_keyword_prompt': 'Keyword for “{label}”',
  'search.keyword_gone': 'The target of “{keyword}” no longer exists; change it under Settings → Search.',
  'search.keyword_added': 'Keyword “{keyword}” added',
  'tools.new': 'New',
  'tools.password': 'Password · {length} characters',
  'tools.password_all': 'Letters, digits and symbols',
  'tools.password_no_symbols': 'Letters and digits',
  'tools.password_digits': 'Digits only',
  'tools.password_note': 'made in your browser and never saved',
  'tools.uuid': 'UUID v{version}',
  'tools.colour': 'Colour',
  'tools.unix_to_date': 'Unix time {seconds}',
  'tools.date_to_unix': 'Unix time',
  'tools.local': 'local',
  'tools.network': 'Network {network}',
  'tools.mask': 'Mask',
  'tools.broadcast': 'Broadcast',
  'tools.first_host': 'First host',
  'tools.last_host': 'Last host',
  'tools.hosts': 'Usable hosts',
  'tools.addresses': 'Addresses',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
