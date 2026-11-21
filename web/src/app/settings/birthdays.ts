import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'birthdays',
  label: 'settings.section.birthdays',
  module: 'birthdays',
  area: 'dashboard',
  load: () => import('../../features/birthdays').then((m) => m.BirthdaysSettings),
  props: (c) => ({ birthdays: c.birthdays, admin: c.auth.admin }),
});
