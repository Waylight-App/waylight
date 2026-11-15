<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** The own e-mail address for reset links (members and guests, #296). */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  let email = $state('');
  let error = $state('');
  let message = $state('');

  $effect(() => {
    auth
      .myEmail()
      .then((e) => (email = e))
      .catch((err) => (error = errorMessage(err)));
  });

  async function save(e: SubmitEvent) {
    e.preventDefault();
    error = message = '';
    try {
      email = await auth.setMyEmail(email);
      message = t('common.saved');
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<form onsubmit={save}>
  <p class="muted">{auth.passwordReset ? t('auth.reset_email_hint') : t('auth.reset_email_unavailable')}</p>
  <label class="field">
    <span>{t('auth.reset_email')} {t('common.optional')}</span>
    <input class="input" type="email" autocomplete="email" bind:value={email} />
  </label>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="muted" role="status">{message}</p>{/if}
  <button class="btn" type="submit">{t('common.save')}</button>
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
</style>
