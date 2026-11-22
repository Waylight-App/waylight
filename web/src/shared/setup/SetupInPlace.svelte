<script lang="ts">
  import type { Snippet } from 'svelte';
  import { t } from '../i18n';
  import EmptyState from '../ui/EmptyState.svelte';
  import SetupAssistant from './SetupAssistant.svelte';
  import type { SavedConnections } from '../connection/connection';
  import { type SetupDescriptor, useMaySetUp } from './setup';

  /**
   * What a widget that is not set up shows (#588, #737): the shared symbol,
   * one sentence and *Connect*, which opens the guided setup in a dialog. Viewers who may not set the
   * integration up (members for instance integrations, guests) only read
   * the sentence. Saved connections of the service come first (#709):
   * choosing one stores it in the widget instead of asking for an address.
   */
  interface Props {
    text: string;
    descriptor: SetupDescriptor;
    /** True for integrations of the instance, which only admins set up. */
    instance: boolean;
    ondone: () => void;
    connections?: SavedConnections;
    /** An action of the widget that comes before connecting, e.g. creating a list. */
    children?: Snippet;
    /** Names the connect button where "Connect" alone is vague. */
    label?: string;
  }
  let { text, descriptor, instance, ondone, connections, children, label }: Props = $props();
  // svelte-ignore state_referenced_locally
  const allowed = useMaySetUp(instance);
  let open = $state(false);
  let saved = $state<{ id: string; name: string }[]>([]);
  // svelte-ignore state_referenced_locally
  if (connections)
    void Promise.all(connections.services.map((s) => connections.load(s)))
      .then((lists) => (saved = lists.flat()))
      .catch(() => undefined);
</script>

{#if open}
  <SetupAssistant
    {descriptor}
    onclose={() => (open = false)}
    ondone={() => {
      open = false;
      ondone();
    }}
  />
{/if}
<EmptyState {text}>
  {#if allowed}
    {@render children?.()}
    {#each saved as c (c.id)}
      <button class="btn btn-primary" type="button" onclick={() => void connections?.choose(c.id)}
        >{t('setup.use_connection', { name: c.name })}</button
      >
    {/each}
    <button
      class={saved.length || children ? 'btn' : 'btn btn-primary'}
      type="button"
      onclick={() => (open = true)}>{label ?? t('setup.connect')}</button
    >
  {/if}
</EmptyState>
