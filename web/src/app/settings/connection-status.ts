import { defineSettingsSection } from '../settingsKit';
import { definitionFor } from '../widgets';
import { integrationName, openSection, sectionFor, tabName } from '../attention.svelte';

// Connections & services (#591): every integration with its state and the
// widgets that use it, next to the saved places.
export const section = defineSettingsSection({
  id: 'connection-status',
  label: 'settings.section.connection_status',
  module: 'attention',
  area: 'connections',
  load: () => import('../../features/attention/ui/ServicesList.svelte').then((m) => m.default),
  props: (c) => ({
    attention: c.attention,
    nameOf: integrationName,
    widgetName: (kind: string) => definitionFor(kind)?.type.title ?? kind,
    tabName,
    sectionOf: sectionFor,
    onopen: openSection,
  }),
});
