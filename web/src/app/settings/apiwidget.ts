import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'apiwidget',
  label: 'settings.section.apiwidget',
  module: 'apiwidget',
  homelab: true,
  // Raw JSON endpoints are for power users (#602).
  advanced: true,
  load: () => import('../../features/apiwidget').then((m) => m.ApiWidgetSettings),
  props: (c) => ({ apiwidget: c.apiwidget }),
});
