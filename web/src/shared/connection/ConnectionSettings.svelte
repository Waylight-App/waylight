<script lang="ts">
  import { t } from '../i18n';
  import { type ChoiceLoader, connectionIdOf, withConnectionChoice } from './connection';
  import ConnectionSelect from './ConnectionSelect.svelte';

  /** Settings of a widget whose only choice is its connection (#709). */
  interface Props {
    services: string[];
    config: Record<string, unknown>;
    connections: ChoiceLoader;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { services, config, connections, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let connectionId = $state(connectionIdOf(config));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    void onsave(withConnectionChoice(config, connectionId));
  }}
>
  <ConnectionSelect {services} load={connections} bind:value={connectionId} always />
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
