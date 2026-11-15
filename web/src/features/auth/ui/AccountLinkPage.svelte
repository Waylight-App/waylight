<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import type { LinkInfo } from '../domain/auth';
  import type { AuthUseCases } from '../usecase/authUseCases';
  import AuthCard from './AuthCard.svelte';

  /**
   * The page behind invitation and reset links (#296). The token came in the
   * URL fragment and was removed from the address bar before this page
   * loaded; it is sent to the server only in request bodies.
   */
  interface Props {
    auth: AuthUseCases;
    token: string;
    /** Leaves the page for the sign-in screen. */
    ondone: () => void;
  }
  let { auth, token, ondone }: Props = $props();

  let info = $state<LinkInfo | null>(null);
  let done = $state(false);
  let password = $state('');
  let confirmation = $state('');
  let error = $state('');
  let busy = $state(false);

  $effect(() => {
    auth
      .inspectLink(token)
      .then((i) => (info = i))
      .catch((err) => (error = errorMessage(err)));
  });

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    busy = true;
    error = '';
    try {
      info = await auth.acceptLink(token, password, confirmation);
      password = confirmation = '';
      done = true;
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }
</script>

<AuthCard
  title={info?.purpose === 'reset' ? t('auth.link_reset_title') : t('auth.link_invite_title')}
  subtitle={info ? t('auth.link_for', { name: info.name }) : ''}
>
  {#if done}
    <p role="status">{info?.purpose === 'reset' ? t('auth.link_reset_done') : t('auth.link_invite_done')}</p>
    <button class="btn btn-primary" type="button" onclick={ondone}>{t('auth.link_to_login')}</button>
  {:else if info}
    <form onsubmit={submit}>
      <!-- The name helps password managers store the new password under it. -->
      <input type="text" name="username" autocomplete="username" value={info.name} hidden readonly />
      <label class="field">
        <span>{t('auth.password_at_least_12_characters')}</span>
        <input
          class="input"
          type="password"
          name="password"
          autocomplete="new-password"
          bind:value={password}
          required
        />
      </label>
      <label class="field">
        <span>{t('auth.repeat_password')}</span>
        <input
          class="input"
          type="password"
          name="confirmation"
          autocomplete="new-password"
          bind:value={confirmation}
          required
        />
      </label>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <button class="btn btn-primary" type="submit" disabled={busy}>{t('auth.link_set_password')}</button>
    </form>
  {:else if error}
    <p class="error" role="alert">{error}</p>
    <button class="btn" type="button" onclick={ondone}>{t('auth.link_to_login')}</button>
  {/if}
</AuthCard>

<style>
  form {
    display: grid;
    gap: var(--space-2);
  }
</style>
