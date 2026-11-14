<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseConfig, ZONES, zoneName } from '../domain/electricity';

  /** Props of the electricity price widget settings: the bidding zone. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let zone = $state(parseConfig(config).zone);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ zone });
  }}
>
  <label class="field"
    ><span>{t('electricity.zone')}</span><select class="input" bind:value={zone}>
      {#each ZONES as z (z)}<option value={z}>{zoneName(z)} ({z})</option>{/each}
    </select></label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
