<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseConfig } from '../domain/contracts';

  /** Props of the contracts widget settings (#687). */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();

  // svelte-ignore state_referenced_locally
  let currency = $state(parseConfig(config).currency);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    await onsave({ ...config, ...parseConfig({ currency: currency.trim().toUpperCase() }) });
  }
</script>

<form onsubmit={submit}>
  <label class="field">
    <span>{t('contracts.currency')}</span>
    <input class="input" bind:value={currency} maxlength="3" autocomplete="off" />
    <span class="muted small">{t('contracts.currency_hint')}</span>
  </label>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  .small {
    font-size: var(--fs-xs);
  }
</style>
