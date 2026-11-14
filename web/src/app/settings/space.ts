import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'space',
  label: 'settings.section.space',
  module: 'space',
  group: 'content',
  load: () => import('../../features/space').then((m) => m.SpaceSettings),
  props: (c) => ({ space: c.space }),
});
