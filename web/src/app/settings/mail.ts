import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'mail',
  label: 'settings.section.mail',
  module: 'mail',
  area: 'dashboard',
  load: () => import('../../features/mail/ui/MailSettings.svelte').then((m) => m.default),
  props: (c) => ({ mail: c.mail }),
});
