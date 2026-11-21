import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'wiki',
  label: 'settings.section.wiki',
  module: 'wiki',
  load: () => import('../../features/wiki').then((m) => m.WikiSettings),
  props: (c) => ({ wiki: c.wiki }),
});
