import { defineSettingsSection } from '../settingsKit';

export const section = defineSettingsSection({
  id: 'todos',
  label: 'settings.section.todos',
  module: 'todos',
  group: 'content',
  load: () => import('../../features/todos/ui/TodoSourcesSettings.svelte').then((m) => m.default),
  props: (c) => ({ sources: c.todoSources }),
});
