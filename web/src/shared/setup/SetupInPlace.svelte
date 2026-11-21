<script lang="ts">
  import { t } from '../i18n';
  import SetupAssistant from './SetupAssistant.svelte';
  import { type SetupDescriptor, useMaySetUp } from './setup';

  /**
   * What a widget that is not set up shows (#588): one sentence and *Set up*,
   * which opens the assistant in the widget. Viewers who may not set the
   * integration up (members for instance integrations, guests) only read
   * the sentence.
   */
  interface Props {
    text: string;
    descriptor: SetupDescriptor;
    /** True for integrations of the instance, which only admins set up. */
    instance: boolean;
    ondone: () => void;
  }
  let { text, descriptor, instance, ondone }: Props = $props();
  // svelte-ignore state_referenced_locally
  const allowed = useMaySetUp(instance);
  let open = $state(false);
</script>

{#if open}
  <SetupAssistant {descriptor} {ondone} />
{:else}
  <div class="hint">
    <p class="muted">{text}</p>
    {#if allowed}
      <button class="btn btn-primary" type="button" onclick={() => (open = true)}>{t('common.set_up')}</button
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
