import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'readlater',
  label: 'settings.section.readlater',
  module: 'readlater',
  area: 'dashboard',
  load: () => import('../../features/readlater').then((m) => m.ReadLaterSettings),
  props: (c) => ({ readlater: c.readlater }),
});
