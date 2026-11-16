<script lang="ts">
  import type { Mountable, PluginContext } from './api';

  /** Hosts a plugin screen: the plugin mounts into this element (ADR 0021). */
  interface Props {
    mountable: Mountable;
    context: PluginContext;
    props?: Record<string, unknown>;
    /** Set by the widget settings dialog; forwarded to the plugin. */
    onsave?: (config: Record<string, unknown>) => void;
  }
  let { mountable, context, props = {}, onsave }: Props = $props();

  let host = $state<HTMLElement>();
  let failed = $state(false);

  $effect(() => {
    if (!host) return;
    let unmount: (() => void) | undefined;
    try {
      unmount = mountable.mount(host, onsave ? { ...props, onsave } : { ...props }, context);
    } catch (err) {
      // A broken plugin must not take the page down with it.
      console.error('plugin failed to mount', err);
      failed = true;
    }
    return () => {
      try {
        unmount?.();
      } catch (err) {
        console.error('plugin failed to unmount', err);
      }
    };
  });
</script>

<div class="plugin" bind:this={host}></div>
{#if failed}<p class="error" role="alert">{context.errorMessage(new Error('plugin'))}</p>{/if}

<style>
  .plugin {
    display: contents;
  }
</style>
