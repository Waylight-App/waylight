<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { MailAccount } from '../domain/mail';
  import type { AccountInput, MailUseCases } from '../usecase/mailUseCases';

  /** Settings → Mail: the IMAP accounts (#283). */
  interface Props {
    mail: MailUseCases;
  }
  let { mail }: Props = $props();

  const empty = (): AccountInput => ({
    name: '',
    host: '',
    port: 993,
    username: '',
    password: '',
    folder: 'INBOX',
    webmailUrl: '',
  });
  let accounts = $state<MailAccount[]>([]);
  let form = $state(empty());
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  async function load() {
    accounts = await mail.accounts();
  }
  $effect(() => {
    void load();
  });

  async function run(action: () => Promise<void>) {
    error = message = '';
    busy = true;
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<p class="muted">{t('mail.intro')}</p>
<p class="muted small">{t('mail.guests')}</p>
<h3>{t('mail.accounts')}</h3>
<ul class="list">
  {#each accounts as a (a.id)}
    <li>
      <span class="what"
        ><strong>{a.name}</strong><span class="muted small">{a.username} · {a.host} · {a.folder}</span></span
      >
      <button
        class="btn btn-icon"
        type="button"
        disabled={busy}
        aria-label={t('mail.remove', { name: a.name })}
        title={t('mail.remove', { name: a.name })}
        onclick={() => run(async () => (await mail.deleteAccount(a.id), await load()))}>✕</button
      >
    </li>
  {:else}
    <li class="muted">{t('mail.no_accounts')}</li>
  {/each}
</ul>
<h3>{t('mail.add')}</h3>
<form
  onsubmit={(e) => {
    e.preventDefault();
    void run(async () => {
      await mail.saveAccount('', form);
      form = empty();
      message = t('mail.saved');
      await load();
    });
  }}
>
  <label class="field"
    ><span>{t('mail.name')}</span><input
      class="input"
      bind:value={form.name}
      maxlength="60"
      required
    /></label
  >
  <div class="row">
    <label class="field grow"
      ><span>{t('mail.host')}</span><input
        class="input"
        bind:value={form.host}
        placeholder="imap.example.com"
        required
      /></label
    >
    <label class="field port"
      ><span>{t('mail.port')}</span><input
        class="input"
        type="number"
        min="1"
        max="65535"
        bind:value={form.port}
      /></label
    >
  </div>
  <label class="field"
    ><span>{t('mail.username')}</span><input
      class="input"
      bind:value={form.username}
      autocomplete="off"
      required
    /></label
  >
  <label class="field"
    ><span>{t('mail.password')}</span><input
      class="input"
      type="password"
      bind:value={form.password}
      autocomplete="new-password"
      required
    /></label
  >
  <label class="field"><span>{t('mail.folder')}</span><input class="input" bind:value={form.folder} /></label>
  <label class="field"
    ><span>{t('mail.webmail')}</span><input
      class="input"
      type="url"
      bind:value={form.webmailUrl}
      placeholder="https://"
    /></label
  >
  <p><button class="btn btn-primary" type="submit" disabled={busy}>{t('mail.save')}</button></p>
</form>
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  .list {
    list-style: none;
    margin: 0 0 var(--space-3);
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .what {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .row {
    display: flex;
    gap: var(--space-2);
  }
  .grow {
    flex: 1;
  }
  .port {
    width: 7rem;
  }
  .small {
    font-size: var(--fs-xs);
  }
  h3 {
    font-size: var(--fs-sm);
    margin: var(--space-3) 0 var(--space-2);
  }
</style>
