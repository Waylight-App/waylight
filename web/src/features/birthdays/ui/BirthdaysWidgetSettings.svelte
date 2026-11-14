<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseConfig } from '../domain/birthdays';

  /** Props of the birthdays widget settings: how many entries to show. */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let limit = $state(parseConfig(config).limit);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ limit: parseConfig({ limit }).limit });
  }}
>
  <label class="field"
    ><span>{t('birthdays.limit')}</span><input
      class="input"
      type="number"
      min="1"
      max="50"
      bind:value={limit}
    /></label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
