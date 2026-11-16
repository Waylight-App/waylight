<script lang="ts">
  import type { PluginContext } from '../../../../web/src/app/plugins/api';
  import { text } from '../lib/texts';
  import { parseConfig, PresenceApi } from './presence';
  import { texts } from './texts';

  /** Settings of the presence widget (#346): which group it shows. */
  interface Props {
    context: PluginContext;
    config?: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => void;
  }
  let { context, config, onsave }: Props = $props();

  const tr = (key: string) => text(texts, key, context.language);
  // The form starts from the saved config and is edited locally.
  // svelte-ignore state_referenced_locally
  let group = $state(parseConfig(config).group);
  let groups = $state<{ id: string; name: string }[]>([]);
  let error = $state('');

  $effect(() => {
    new PresenceApi(context.http)
      .groups()
      .then((list) => (groups = list))
      .catch((err) => (error = context.errorMessage(err)));
  });

  function submit(e: SubmitEvent) {
    e.preventDefault();
    onsave({ group });
  }
</script>

<form onsubmit={submit}>
  <label class="field"
    ><span>{tr('group')}</span>
    <select class="input" bind:value={group}>
      <option value="">{tr('everyone')}</option>
      {#each groups as g (g.id)}<option value={g.id}>{g.name}</option>{/each}
    </select></label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p><button class="btn btn-primary" type="submit">{tr('save')}</button></p>
</form>
