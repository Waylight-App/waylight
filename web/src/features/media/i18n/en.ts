/** English UI texts of the media feature (#193). */
export const en = {
  'widget.media.title': 'Media server',
  'widget.media.description': 'Now playing and recently added on Jellyfin or Plex.',
  'settings.section.media': 'Media server',
  'media.connect_under_settings': 'Connect Jellyfin or Plex under Settings → Media server.',
  'media.now_playing': 'Now playing',
  'media.nothing_playing': 'Nothing is playing.',
  'media.recently_added': 'Recently added',
  'media.server': 'Media server',
  'media.url': '{name} address',
  'media.jellyfin_key': 'Access key {value}',
  'media.plex_token': 'Plex token {value}',
  'media.jellyfin_hint':
    'Create an API key in Jellyfin under Dashboard → API keys. It is stored encrypted; posters are loaded through Waylight.',
  'media.plex_hint':
    'Use the X-Plex-Token of your account (Plex support article "Finding an authentication token"). It is stored encrypted; posters are loaded through Waylight.',
  'media.setup_token': 'Access key or token',
  'media.setup_step_jellyfin': 'Jellyfin: open Dashboard → API Keys, press + and name the key Waylight.',
  'media.setup_step_plex':
    'Plex: open any item in Plex Web, choose ⋯ → Get Info → View XML; the token is the value after X-Plex-Token= in the address bar.',
  'media.setup_step_copy': 'Copy the key or token and paste it below.',
  'media.setup_found': '{recent} recently added, {playing} playing now',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
