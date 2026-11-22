import { defineSettingsSection } from '../settingsKit';
import { definitionFor } from '../widgets';

// Named connections (#708): address and access of a service, saved once and
// chosen by every widget of that service. The module holds instance data, so
// members and guests do not see the section; the API refuses them as well.
export const section = defineSettingsSection({
  id: 'integrations',
  label: 'settings.section.integrations',
  module: 'integrations',
  keywords: ['connections', 'services', 'access', 'key', 'verbindungen', 'dienste', 'zugang', 'schlüssel'],
  load: () => import('../../features/integrations').then((m) => m.ConnectionsManager),
  props: (c) => ({
    connections: c.integrations,
    widgetName: (kind: string) => definitionFor(kind)?.type.title ?? kind,
    tabs: () => c.dashboard.load(),
  }),
});
