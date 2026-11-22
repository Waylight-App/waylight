<script lang="ts">
  import type { ChoiceLoader } from '../../../shared/connection/connection';
  import ConnectionSelect from '../../../shared/connection/ConnectionSelect.svelte';
  import { t } from '../../../shared/i18n';
  import { dnsProviders, parseWidgetConfig, providerName } from '../domain/adguard';

  /** Props of the DNS filter widget settings: which provider it shows (#83). */
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
  <label class="field">
    <span>{t('adguard.provider')}</span>
    <select class="input" bind:value={form.provider} onchange={() => (form.connectionId = '')}>
      {#each dnsProviders as p (p)}<option value={p}>{providerName(p)}</option>{/each}
    </select>
  </label>
  {#key form.provider}
    <ConnectionSelect services={[form.provider]} load={connections} bind:value={form.connectionId} />
  {/key}
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
