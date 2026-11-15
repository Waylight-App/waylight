<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /**
   * "Forgot your password?" on the sign-in page (#296). The answer is the
   * same whether or not the name exists, so the page never says more.
   */
  interface Props {
    auth: AuthUseCases;
    onback: () => void;
  }
  let { auth, onback }: Props = $props();

  let name = $state('');
  let sent = $state(false);
  let error = $state('');
  let busy = $state(false);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    try {
      await auth.requestPasswordReset(name);
      sent = true;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if sent}
  <p role="status">{t('auth.reset_mail_sent')}</p>
{:else}
  <form onsubmit={submit}>
    <p class="muted">{t('auth.reset_mail_hint')}</p>
    <label class="field">
      <span>{t('auth.guest_name')}</span>
      <input
        class="input"
        name="username"
        autocomplete="username"
        autocapitalize="none"
        bind:value={name}
        required
      />
    </label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit" disabled={busy}>{t('auth.reset_mail_send')}</button>
  </form>
{/if}
<button class="btn btn-icon" type="button" onclick={onback}>{t('auth.back')}</button>

<style>
  form {
    display: grid;
    gap: var(--space-2);
  }
</style>
