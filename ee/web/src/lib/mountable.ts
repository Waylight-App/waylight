import { mount, unmount, type Component } from 'svelte';
import type { Mountable, PluginContext } from '../../../../web/src/app/plugins/api';

/**
 * Turns a Svelte component of the plugin into a Mountable (ADR 0021): the
 * core hands over a host element, the plugin mounts its own component tree
 * into it with its own Svelte runtime.
 */
export function mountable<P extends Record<string, unknown>>(
  component: Component<P & { context: PluginContext }>,
): Mountable {
  return {
    mount(host, props, context) {
      const app = mount(component, {
        target: host,
        props: { ...(props as P), context },
      });
      return () => void unmount(app);
    },
  };
}

/**
 * Like mountable(), but loads the component on first use, so heavy screens
 * (Markdown, editors) do not weigh on plugin.js and land in their own chunk.
 */
export function lazyMountable<P extends Record<string, unknown>>(
  load: () => Promise<{ default: Component<P & { context: PluginContext }> }>,
): Mountable {
  return {
    mount(host, props, context) {
      let app: Record<string, unknown> | undefined;
      let gone = false;
      load().then(
        ({ default: component }) => {
          if (!gone)
            app = mount(component, {
              target: host,
              props: { ...(props as P), context },
            });
        },
        (err: unknown) => console.error('plugin screen failed to load', err),
      );
      return () => {
        gone = true;
        if (app) void unmount(app);
      };
    },
  };
}
