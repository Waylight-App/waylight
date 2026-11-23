<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import type { BirthdaysUseCases } from '../usecase/birthdaysUseCases';

  /** Settings → Birthdays (#287): vCard import, CardDAV address book and reminder. */
  interface Props {
    birthdays: BirthdaysUseCases;
    /**
     * The reminder goes to the instance's notification channels, so the
     * server lets only admins switch it on (#382); members don't see it (#387).
     */
    admin?: boolean;
  }
  let { birthdays, admin = true }: Props = $props();

  let remindEnabled = $state(false);
  let remindAt = $state('08:00');
  let cardDavUrl = $state('');
  let username = $state('');
  let password = $state('');
  let hasPassword = $state(false);
  let error = $state('');
  let message = $state('');
  let importMessage = $state('');
  let importError = $state('');
  let busy = $state(false);

  $effect(() => {
    birthdays
      .settings()
      .then((s) => {
        ({ remindEnabled, remindAt, cardDavUrl, username, hasPassword } = s);
      })
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      const s = await birthdays.saveSettings({ remindEnabled, remindAt, cardDavUrl, username, password });
      ({ cardDavUrl, username, hasPassword } = s);
      password = '';
      message = t('birthdays.saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function importFile(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    importMessage = importError = '';
    try {
      const r = await birthdays.importVCard(file);
      importMessage = t('birthdays.imported', { added: r.added, duplicates: r.duplicates });
    } catch (err) {
      importError = errorMessage(err);
    } finally {
      input.value = '';
    }
  }
</script>

<h3>{t('birthdays.import_title')}</h3>
<p class="muted small">{t('birthdays.import_hint')}</p>
<label class="btn file"
  >{t('birthdays.import_button')}<input
    type="file"
    accept=".vcf,text/vcard,text/x-vcard"
    onchange={importFile}
  /></label
>
{#if importError}<p class="error" role="alert">{importError}</p>{/if}
{#if importMessage}<p class="muted" role="status">{importMessage}</p>{/if}

<form onsubmit={submit}>
  <h3>{t('birthdays.book_title')}</h3>
  <p class="muted small">{t('birthdays.book_hint')}</p>
  <label class="field"
    ><span>{t('birthdays.book_url')} {t('common.optional')}</span><input
      class="input"
      type="url"
      bind:value={cardDavUrl}
      placeholder="https://cloud.example.org/remote.php/dav/addressbooks/users/me/contacts/"
    /></label
  >
  {#if cardDavUrl.trim()}
    <label class="field"
      ><span>{t('birthdays.username')}</span><input
        class="input"
        autocomplete="username"
        bind:value={username}
      /></label
    >
    <label class="field"
      ><span>{t('birthdays.password')}</span><input
        class="input"
        type="password"
        autocomplete="off"
        bind:value={password}
        placeholder={hasPassword ? t('birthdays.password_stored') : ''}
      /></label
    >
  {/if}

  {#if admin}
    <h3>{t('birthdays.remind_title')}</h3>
    <label class="check"
      ><input type="checkbox" bind:checked={remindEnabled} /> {t('birthdays.remind_on')}</label
    >
    {#if remindEnabled}
      <label class="field time"
        ><span>{t('birthdays.remind_at')}</span><input
          class="input"
          type="time"
          bind:value={remindAt}
          required
        /></label
      >
    {/if}
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted" role="status">{message}</p>{/if}
  <button class="btn btn-primary" type="submit" disabled={busy}>{t('common.save')}</button>
</form>

<style>
  form {
    display: grid;
    gap: var(--space-2);
    justify-items: start;
  }
  form .field {
    justify-self: stretch;
  }
  form .time {
    justify-self: start;
  }
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  h3 {
    margin: var(--space-3) 0 var(--space-1);
    font-size: inherit;
  }
  .small {
    font-size: var(--fs-xs);
    margin: 0 0 var(--space-1);
  }
  .file {
    position: relative;
    overflow: hidden;
  }
  .file input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }
</style>
