<script lang="ts">
  import type { ChoiceLoader } from '../../../shared/connection/connection';
  import ConnectionSelect from '../../../shared/connection/ConnectionSelect.svelte';
  import { t } from '../../../shared/i18n';
  import { parseWidgetConfig } from '../domain/proxmox';

  /** Props of the Proxmox widget settings. */
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
  <ConnectionSelect services={['proxmox']} load={connections} bind:value={form.connectionId} />
  <label><input type="checkbox" bind:checked={form.runningOnly} /> {t('proxmox.running_only')}</label>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
