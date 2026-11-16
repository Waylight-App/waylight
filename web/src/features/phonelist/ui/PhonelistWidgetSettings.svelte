<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parsePhonelistConfig } from '../domain/phonelist';

  /** Props of the phone list widget settings (#417). */
  interface Props {
    config: Record<string, unknown>;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parsePhonelistConfig(config));
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...config, ...form });
  }}
>
  <label><input type="checkbox" bind:checked={form.groupByTeam} /> {t('phonelist.group_by_team')}</label>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
