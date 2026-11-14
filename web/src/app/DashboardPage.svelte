<script lang="ts">
  import { DashboardView, type Tab, type Widget, type WidgetConfig } from '../features/dashboard';
  import type { Container } from './container';
  import LazyComponent from './LazyComponent.svelte';
  import { definitionFor, iconFor, widgetDefinitions } from './widgets';
  import { categoryShown, type StartProfile } from '../features/appearance/domain/appearance';

  /** Props of the dashboard page. */
  interface Props {
    container: Container;
    editing: boolean;
    /** What the start page shows (#281). */
    profile?: StartProfile;
    /** Widget kinds enabled in the backend. */
    widgetKinds: Set<string>;
    /** Element id in the top bar that hosts the tab bar (#167). */
    tabBarTarget?: string;
    /** Reports tabs and the active tab to the shell. */
    onstate?: (tabs: Tab[], activeId: string) => void;
  }
  let { container, editing, profile = 'homelab', widgetKinds, tabBarTarget, onstate }: Props = $props();

  let view = $state<{ selectTabAt: (index: number) => void; undo: () => void; redo: () => void }>();

  /** Selects the n-th tab (0-based). */
  export function selectTabAt(index: number) {
    view?.selectTabAt(index);
  }

  /** Undo/redo in edit mode (#93). */
  export function undo() {
    view?.undo();
  }
  export function redo() {
    view?.redo();
  }

  // The "everyday" profile hides home-server and network widgets from the
  // picker (#281); widgets already placed stay.
  const types = $derived(
    widgetDefinitions
      .filter((d) => widgetKinds.has(d.type.kind) && categoryShown(profile, d.type.category))
      .map((d) => d.type),
  );
</script>

<DashboardView
  bind:this={view}
  uc={container.dashboard}
  {editing}
  widgetTypes={types}
  widgetIcon={iconFor}
  hasSettings={(kind) => !!definitionFor(kind)?.settings}
  chromeless={(kind) => !!definitionFor(kind)?.chromeless}
  {tabBarTarget}
  {onstate}
>
  {#snippet content(widget: Widget)}
    {@const def = definitionFor(widget.kind)}
    {#if def}
      <LazyComponent
        renderer={def.view}
        props={def.view.props(widget.config, container)}
        scheduler={container.refresh}
        label={def.type.title}
      />
    {/if}
  {/snippet}
  {#snippet settings(widget: Widget, save: (config: WidgetConfig) => Promise<void>)}
    {@const def = definitionFor(widget.kind)}
    {#if def?.settings}
      <LazyComponent
        renderer={def.settings}
        props={{ ...def.settings.props(widget.config, container), onsave: save }}
        scheduler={container.refresh}
        showStatus={false}
      />
    {/if}
  {/snippet}
</DashboardView>
