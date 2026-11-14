import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'smarthome',
  label: 'settings.section.smarthome',
  module: 'smarthome',
  homelab: true,
  load: () => import('../../features/smarthome').then((m) => m.SmartHomeSettings),
  props: (c) => ({ smarthome: c.smarthome }),
});
