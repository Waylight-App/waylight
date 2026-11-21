import { defineSettingsSection } from '../settingsKit';
import { attention, integrationName, runAttentionAction, tabName } from '../attention.svelte';

// The needs-your-attention page (#592). It belongs to the connections area;
// the settings start page (#600) links it from its attention hint.
export const section = defineSettingsSection({
  id: 'attention',
  label: 'settings.section.attention',
  module: 'attention',
  area: 'connections',
  load: () => import('../../features/attention').then((m) => m.AttentionPage),
  props: () => ({ state: attention, nameOf: integrationName, tabName, onaction: runAttentionAction }),
});
