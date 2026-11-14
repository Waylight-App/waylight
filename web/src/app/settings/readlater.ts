import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'readlater',
  label: 'settings.section.readlater',
  module: 'readlater',
  group: 'content',
  load: () => import('../../features/readlater').then((m) => m.ReadLaterSettings),
  props: (c) => ({ readlater: c.readlater }),
});
