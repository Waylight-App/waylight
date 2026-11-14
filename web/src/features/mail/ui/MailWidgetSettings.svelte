<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { type MailAccount, parseConfig } from '../domain/mail';
  import type { MailUseCases } from '../usecase/mailUseCases';

  /** Props of the mail widget settings: which account. */
  interface Props {
    config: Record<string, unknown>;
    mail: MailUseCases;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, mail, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let account = $state(parseConfig(config).account);
  let accounts = $state<MailAccount[]>([]);
  $effect(() => {
    void mail.accounts().then((a) => (accounts = a));
  });
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave(account ? { account } : {});
  }}
>
  <label class="field"
    ><span>{t('mail.account')}</span><select class="input" bind:value={account}>
      <option value="">{t('mail.all_accounts')}</option>
      {#each accounts as a (a.id)}<option value={a.id}>{a.name}</option>{/each}
    </select></label
  >
  <p class="muted small">{t('mail.guests')}</p>
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>

<style>
  .small {
    font-size: max(0.8rem, 0.85em);
  }
</style>
