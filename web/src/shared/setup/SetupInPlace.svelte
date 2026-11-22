<script lang="ts">
  import { t } from '../i18n';
  import SetupAssistant from './SetupAssistant.svelte';
  import type { SavedConnections } from '../connection/connection';
  import { type SetupDescriptor, useMaySetUp } from './setup';

  /**
   * What a widget that is not set up shows (#588): one sentence and *Set up*,
   * which opens the assistant in the widget. Viewers who may not set the
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
  }
  let { text, descriptor, instance, ondone, connections }: Props = $props();
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
  <SetupAssistant {descriptor} {ondone} />
{:else}
  <div class="hint">
    <p class="muted">{text}</p>
    {#if allowed}
      {#each saved as c (c.id)}
        <button class="btn btn-primary" type="button" onclick={() => void connections?.choose(c.id)}
          >{t('setup.use_connection', { name: c.name })}</button
        >
      {/each}
      <button class={saved.length ? 'btn' : 'btn btn-primary'} type="button" onclick={() => (open = true)}
        >{t('common.set_up')}</button
      >
    {/if}
  </div>
{/if}

<style>
  .hint {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    height: 100%;
    text-align: center;
  }
  .hint p {
    margin: 0;
    max-width: 30ch;
  }
</style>
