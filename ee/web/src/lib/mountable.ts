import { mount, unmount, type Component } from "svelte";
import type {
  Mountable,
  PluginContext,
} from "../../../../web/src/app/plugins/api";

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
