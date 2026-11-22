<script lang="ts">
  import type { ChoiceLoader } from '../../../shared/connection/connection';
  import ConnectionSelect from '../../../shared/connection/ConnectionSelect.svelte';
  import { t } from '../../../shared/i18n';
  import { parseWidgetConfig } from '../domain/arr';

  /** Props of the *arr widget settings: which sections to show. */
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
  <ConnectionSelect
    services={['sonarr', 'radarr', 'prowlarr']}
    load={connections}
    bind:values={form.connectionIds}
  />
  <label class="check"><input type="checkbox" bind:checked={form.health} /> {t('arr.health')}</label>
  <label class="check"><input type="checkbox" bind:checked={form.queue} /> {t('arr.queue')}</label>
  <label class="check"><input type="checkbox" bind:checked={form.upcoming} /> {t('arr.upcoming')}</label>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  .check {
    display: block;
  }
</style>
