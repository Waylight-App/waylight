<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { appCallback, type AppHandover } from '../domain/appHandover';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /** Props of the confirmation of a sign-in for the app (#560). */
  interface Props {
    auth: AuthUseCases;
    handover: AppHandover;
    /** Called once the app was answered, to show the dashboard. */
    ondone: () => void;
  }
  let { auth, handover, ondone }: Props = $props();

  let busy = $state(false);
  let error = $state('');

  // The person confirms every handover: another app could have opened this
  // page, and the confirmation makes sure nobody is signed in unnoticed.
  async function confirm() {
    busy = true;
    error = '';
    try {
      const { code } = await auth.newHandoverCode(handover.challenge);
      location.href = appCallback(handover, { code });
      ondone();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function cancel() {
    location.href = appCallback(handover, { error: 'access_denied' });
    ondone();
  }
</script>

<main class="handover">
  <h1>{t('auth.handover_title')}</h1>
  <p>{auth.username ? t('auth.handover_text_named', { name: auth.username }) : t('auth.handover_text')}</p>
  <p class="muted">{t('auth.handover_hint')}</p>
  <div class="actions">
    <button class="btn btn-primary" type="button" disabled={busy} onclick={confirm}
      >{t('auth.handover_confirm')}</button
    >
    <button class="btn" type="button" disabled={busy} onclick={cancel}>{t('auth.handover_cancel')}</button>
  </div>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</main>

<style>
  .handover {
    max-width: 28rem;
    margin: 15vh auto 0;
    padding: 0 var(--space-3);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
</style>
