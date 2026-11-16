/** English UI texts of the search feature that live with it (#477, #478). */
export const en = {
  'search.own_keywords': 'Own keywords',
  'search.own_keywords_hint':
    'Type a keyword alone in the search field and press Enter to open what it points to, for example “nas” for your NAS. Keywords work on every device you sign in on.',
  'search.own_keyword': 'Keyword',
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
  'search.keyword_added': 'Keyword “{keyword}” added',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
