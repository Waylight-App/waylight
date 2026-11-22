<script lang="ts">
  import type { ChoiceLoader } from '../../../shared/connection/connection';
  import ConnectionSelect from '../../../shared/connection/ConnectionSelect.svelte';
  import { t } from '../../../shared/i18n';
  import { parseWidgetConfig } from '../domain/fuel';

  /** Props of the fuel widget settings: cheapest stations or favourites. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
    connections: ChoiceLoader;
  }
  let { config, onsave, connections }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parseWidgetConfig(config));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <ConnectionSelect services={['tankerkoenig']} load={connections} bind:value={form.connectionId} />
  <label class="field"
    ><span>{t('fuel.view')}</span><select class="input" bind:value={form.view}>
      <option value="cheapest">{t('fuel.view_cheapest')}</option>
      <option value="favourites">{t('fuel.view_favourites')}</option>
    </select></label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
