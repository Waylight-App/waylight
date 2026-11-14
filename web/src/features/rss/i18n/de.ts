import type { en } from './en';

/** German UI texts added to the rss feature; the type makes every key mandatory. */
export const de: Record<keyof typeof en, string> = {
  'rss.catalogue': 'Feed-Katalog',
  'rss.catalogue_browse': 'Katalog durchsuchen',
  'rss.catalogue_search': 'Quellen suchen …',
  'rss.catalogue_language': 'Sprache',
  'rss.catalogue_all_languages': 'Alle Sprachen',
  'rss.catalogue_add': '{title} abonnieren',
  'rss.catalogue_subscribed': 'abonniert',
  'rss.catalogue_none': 'Keine Quelle passt.',
  'rss.topic_news': 'Nachrichten',
  'rss.topic_tech': 'Technik',
  'rss.topic_selfhosting': 'Self-Hosting und Homelab',
  'rss.topic_science': 'Wissenschaft',
  'rss.topic_sports': 'Sport',
  'rss.topic_fun': 'Unterhaltung',
  'rss.compact': 'Kompakt (nur Titel)',
  'rss.opml_import': 'OPML importieren …',
  'rss.opml_export': 'OPML exportieren',
  'rss.opml_report': '{added} Feeds hinzugefügt, {duplicates} schon abonniert, {invalid} nicht importiert.',
  'rss.opml_too_large': 'Die OPML-Datei ist größer als 2 MB.',
};
