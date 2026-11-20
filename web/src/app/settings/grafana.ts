import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'grafana',
  label: 'settings.section.grafana',
  module: 'grafana',
  homelab: true,
  load: () => import('../../features/grafana').then((m) => m.GrafanaSettings),
  props: (c) => ({ grafana: c.grafana }),
});
