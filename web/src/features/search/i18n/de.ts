import type { en } from './en';

/** German UI texts of the search feature that live with it (#477, #478). */
export const de: Record<keyof typeof en, string> = {
  'search.own_keywords': 'Eigene Stichwörter',
  'search.own_keywords_hint':
    'Tippe ein Stichwort allein ins Suchfeld und drücke Enter, um zu öffnen, worauf es zeigt, zum Beispiel „nas“ für dein NAS. Stichwörter gelten auf jedem Gerät, auf dem du angemeldet bist.',
  'search.own_keyword': 'Stichwort',
  'search.keyword_target_kind': 'Öffnet',
  'search.keyword_target': 'Ziel',
  'search.keyword_pick': 'Auswählen…',
  'search.keyword_query_placeholder': 'Suche, z. B. Müll nächste Woche',
  'search.kw_service': 'Dienst-Kachel',
  'search.kw_bookmark': 'Lesezeichen',
  'search.kw_widget': 'Widget',
  'search.kw_url': 'Webadresse',
  'search.kw_query': 'Suche',
  'search.delete_keyword': 'Stichwort {keyword} löschen',
  'search.keyword_result': 'Dein Stichwort',
  'search.add_keyword': 'Stichwort hinzufügen…',
  'search.add_keyword_prompt': 'Stichwort für „{label}“',
  'search.keyword_added': 'Stichwort „{keyword}“ hinzugefügt',
};
