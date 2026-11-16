<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import type { PhonelistSettings } from '../domain/phonelist';
  import type { PhonelistUseCases } from '../usecase/phonelistUseCases';

  /** Settings → Phone list (#300): the CardDAV address book of the instance. */
  interface Props {
    phonelist: PhonelistUseCases;
  }
  let { phonelist }: Props = $props();

  let current = $state<PhonelistSettings | null>(null);
  let cardDavUrl = $state('');
  let username = $state('');
  let password = $state('');
  let error = $state('');
  let message = $state('');
  let busy = $state(false);

  function show(s: PhonelistSettings) {
    current = s;
    cardDavUrl = s.cardDavUrl;
    username = s.username;
    password = '';
  }

  $effect(() => {
    phonelist
      .settings()
      .then(show)
      .catch((err) => (error = errorMessage(err)));
  });

  async function save(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    busy = true;
    try {
      show(await phonelist.saveSettings({ cardDavUrl, username, password }));
      message = t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<form onsubmit={save}>
  <h3>{t('phonelist.book_title')}</h3>
  <p class="muted small">{t('phonelist.book_hint')}</p>
  <label class="field"
    ><span>{t('phonelist.book_url')} {t('common.optional')}</span><input
      class="input"
      type="url"
      bind:value={cardDavUrl}
      placeholder="https://cloud.example.org/remote.php/dav/addressbooks/users/office/staff/"
    /></label
  >
  {#if cardDavUrl.trim()}
    <label class="field"
      ><span>{t('phonelist.book_username')}</span><input
        class="input"
        autocomplete="off"
        bind:value={username}
      /></label
    >
    <label class="field"
      ><span>{t('phonelist.book_password')}</span><input
        class="input"
        type="password"
        autocomplete="off"
        bind:value={password}
        placeholder={current?.hasPassword ? t('phonelist.book_password_stored') : ''}
      /></label
    >
  {/if}
  {#if current?.lastSync}
    <p class="muted small">
      {t('phonelist.book_last_sync', { time: current.lastSync.toLocaleString(intlLocale()) })}
    </p>
  {/if}
  {#if current?.lastError}<p class="error" role="alert">
      {t('phonelist.book_error', { error: current.lastError })}
    </p>{/if}
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
  h3 {
    margin: 0;
    font-size: 1em;
  }
  .small {
    font-size: 0.85em;
    margin: 0;
  }
</style>
