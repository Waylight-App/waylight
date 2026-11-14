/** English UI texts added to the rss feature (#125, #193). */
export const en = {
  'rss.catalogue': 'Feed catalogue',
  'rss.catalogue_browse': 'Browse catalogue',
  'rss.catalogue_search': 'Search sources…',
  'rss.catalogue_language': 'Language',
  'rss.catalogue_all_languages': 'All languages',
  'rss.catalogue_add': 'Subscribe to {title}',
  'rss.catalogue_subscribed': 'subscribed',
  'rss.catalogue_none': 'No source matches.',
  'rss.topic_news': 'News',
  'rss.topic_tech': 'Technology',
  'rss.topic_selfhosting': 'Self-hosting and home lab',
  'rss.topic_science': 'Science',
  'rss.topic_sports': 'Sports',
  'rss.topic_fun': 'Fun',
  'rss.compact': 'Compact (titles only)',
  'rss.opml_import': 'Import OPML…',
  'rss.opml_export': 'Export OPML',
  'rss.opml_report': '{added} feeds added, {duplicates} already subscribed, {invalid} not imported.',
  'rss.opml_too_large': 'The OPML file is larger than 2 MB.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
