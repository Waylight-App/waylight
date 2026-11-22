/** English UI texts of the smarthome feature (#193). */
export const en = {
  'widget.smarthome.title': 'Smart home',
  'widget.smarthome.description':
    'States of Home Assistant devices and sensors; optional switches and buttons.',
  'settings.section.smarthome': 'Home Assistant',
  'smarthome.connect_under_settings': 'Connect Home Assistant under Settings → Home Assistant.',
  'smarthome.choose_entities': 'Choose entities in the widget settings.',
  'smarthome.unavailable': 'unavailable',
  'smarthome.on': 'on',
  'smarthome.off': 'off',
  'smarthome.turn_on': 'Turn on',
  'smarthome.turn_off': 'Turn off',
  'smarthome.press': 'Press',
  'smarthome.confirm_toggle': 'Switch {name}?',
  'smarthome.confirm_press': 'Press {name}?',
  'smarthome.search_entities': 'Search entities',
  'smarthome.selected': '{count} of at most {max} selected',
  'smarthome.url': 'Home Assistant address',
  'smarthome.token': 'Long-lived access token {value}',
  'smarthome.token_hint':
    'Create the token in Home Assistant under your profile → Security → Long-lived access tokens. It is stored encrypted and never shown again.',
  'smarthome.actions_hint':
    'Switches and buttons stay read-only unless SMARTHOME_ACTIONS_ENTITIES lists their entity IDs.',
  'smarthome.setup_step_profile': 'Open Home Assistant and click your user name at the bottom left.',
  'smarthome.setup_step_security': 'Open the Security tab and scroll down to Long-lived access tokens.',
  'smarthome.setup_step_create': 'Press Create token, name it Waylight and press OK.',
  'smarthome.setup_step_copy': 'Copy the token and paste it below; Home Assistant shows it only once.',
  'smarthome.setup_token': 'Access token',
  'smarthome.setup_entities': '{n} entities found',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
