/** English UI texts of the immich feature (#193). */
export const en = {
  'widget.immich.title': 'Immich',
  'widget.immich.description': "Photo library statistics and today's memories.",
  'settings.section.immich': 'Immich',
  'immich.connect_under_settings': 'Connect Immich under Settings → Immich.',
  'immich.photos': 'photos',
  'immich.videos': 'videos',
  'immich.storage': 'Storage',
  'immich.years_ago': '{n} years ago today',
  'immich.memory_photo': 'Photo from {n} years ago',
  'immich.show_memories': "Show today's memories",
  'immich.url': 'Immich URL',
  'immich.api_key': 'API key {value}',
  'immich.key_hint':
    'Create an API key under Account settings → API keys with the permissions asset.statistics, server.storage, memory.read and asset.view. It is stored encrypted; photos are loaded through Waylight.',
  'immich.setup_step_profile': 'Open Immich and click your profile picture at the top right.',
  'immich.setup_step_keys': 'Choose Account Settings → API Keys → New API Key.',
  'immich.setup_step_permissions':
    'Name it Waylight, tick asset.statistics, server.storage, memory.read and asset.view, and press Create.',
  'immich.setup_step_copy': 'Copy the key and paste it below.',
  'immich.setup_key': 'API key',
  'immich.setup_photos': '{n} photos, {videos} videos',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
