import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'device',
  label: 'settings.section.device',
  module: 'device',
  area: 'system',
  load: () => import('../../features/device/ui/DeviceSettings.svelte').then((m) => m.default),
  props: (c) => ({ device: c.device }),
});
