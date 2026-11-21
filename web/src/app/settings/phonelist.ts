import { defineSettingsSection } from '../settingsKit';

// The address book of the phone list (#300); instance data, so only admins
// see the section (members only get settings of personal modules).
export const section = defineSettingsSection({
  id: 'phonelist',
  label: 'settings.section.phonelist',
  module: 'phonelist',
  area: 'dashboard',
  load: () => import('../../features/phonelist/ui/PhonelistSettings.svelte').then((m) => m.default),
  props: (c) => ({ phonelist: c.phonelist }),
});
