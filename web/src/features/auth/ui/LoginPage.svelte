<script lang="ts">
  import { untrack } from 'svelte';
  import { t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { redirectToProvider, ssoErrorKey } from '../domain/auth';
  import SsoIcon from './SsoIcon.svelte';
  import type { SecondFactorChallenge } from '../domain/passkey';
  import type { AuthUseCases } from '../usecase/authUseCases';
  import AuthCard from './AuthCard.svelte';
  import ForgotPassword from './ForgotPassword.svelte';

  /** Props of the login screen. */
  interface Props {
    auth: AuthUseCases;
    /** Called after a successful login. */
    ondone: () => void;
    /** The organisation's name (#305); the product name by default. */
    brandName?: string;
    /** Shows a small "powered by" line under a custom brand. */
    poweredBy?: boolean;
    /** The organisation's logos (#305, #513); the product mark without them. */
    logo?: string;
    logoDark?: string;
  }
  let { auth, ondone, brandName = 'trisse-home', poweredBy = false, logo, logoDark }: Props = $props();

  /** Moves the focus to a message that appears, so it is read out at once. */
  function focusOnShow(node: HTMLElement) {
    node.focus();
  }

  const params = new URLSearchParams(location.search);
  // OIDC_AUTO_REDIRECT (#514): without a password form the page has nothing
  // to offer but the provider; ?local=1 keeps it for recovery.
  if (
    redirectToProvider(
      untrack(() => auth.sso),
      params,
    )
  )
    location.assign('/api/auth/oidc/start');
  // The ticket is read once when the page opens; later changes come from submit().
  let ticket = $state(params.get('sso_ticket') ?? untrack(() => auth.ssoTicket));
  const ssoError = ssoErrorKey(params.get('sso_error'));
  if (params.has('sso_ticket') || params.has('sso_error')) history.replaceState(null, '', location.pathname);
  let password = $state('');
  let username = $state('');
  /** Shows the name field for guest accounts. */
  let asGuest = $state(false);
  let code = $state('');
  /** True once the server asked for the second factor. */
  let needsCode = $state(false);
  /** Passkey challenge of the second login step. */
  let challenge = $state<SecondFactorChallenge | null>(null);
  /** Shows the "forgot your password?" form (#296). */
  let forgot = $state(false);

  async function usePasskey() {
    if (!challenge) return;
    busy = true;
    error = '';
    try {
      await auth.loginWithPasskey(challenge);
      password = code = '';
      ondone();
    } catch (err) {
      // The ticket is single use: start again with the password.
      error = errorMessage(err);
      challenge = null;
      needsCode = false;
    } finally {
      busy = false;
    }
  }
  const passwordHidden = $derived(needsCode || !!ticket || (!auth.sso.localLogin && !asGuest));
  let error = $state('');
  let busy = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    busy = true;
    error = '';
    try {
      if (ticket) {
        await auth.completeSso(ticket, code);
        ondone();
        return;
      }
      const second = await auth.login(password, needsCode ? code : undefined, asGuest ? username : undefined);
      if (second) {
        challenge = second;
        needsCode = true;
        return;
      }
      password = code = '';
      ondone();
    } catch (err) {
      if (err instanceof ApiError && err.field === 'totpCode' && !needsCode) {
        needsCode = true;
      } else {
        error = errorMessage(err);
      }
    } finally {
      busy = false;
    }
  }
</script>

<AuthCard
  title={t('auth.welcome_back')}
  subtitle={needsCode ? t('auth.enter_code_hint') : t('auth.please_log_in')}
  brand={{ name: brandName, logo, logoDark }}
>
  {#if ssoError}<p class="error" role="alert" tabindex="-1" use:focusOnShow>{t(ssoError)}</p>{/if}
  {#if forgot}
    <ForgotPassword {auth} onback={() => (forgot = false)} />
  {:else}
    {#if auth.sso.oidc && !ticket && !needsCode}
      <a class="btn btn-primary wide" href="/api/auth/oidc/start"
        >{#if auth.sso.buttonIcon}<SsoIcon icon={auth.sso.buttonIcon} />{/if}{auth.sso.buttonLabel
          ? t('auth.sign_in_with_provider', { name: auth.sso.buttonLabel })
          : t('auth.sign_in_with_sso')}</a
      >
      {#if !passwordHidden}
        <p class="divider"><span>{t('auth.or_with_password')}</span></p>
      {/if}
    {/if}
    <form onsubmit={submit}>
      {#if asGuest}
        <label class="field" class:hidden={needsCode}>
          <span>{t('auth.guest_name')}</span>
          <input
            class="input"
            name="username"
            autocomplete="username"
            autocapitalize="none"
            bind:value={username}
            required
          />
        </label>
      {/if}
      <label class="field" class:hidden={passwordHidden}>
        <span>{t('auth.password')}</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          class="input"
          type="password"
          name="password"
          autocomplete="current-password"
          bind:value={password}
          autofocus
          required={!passwordHidden}
        />
      </label>
      {#if challenge}
        <button class="btn btn-primary" type="button" disabled={busy} onclick={usePasskey}
          >{t('auth.use_passkey')}</button
        >
        <p class="muted">{t('auth.or_enter_code')}</p>
      {/if}
      {#if needsCode || ticket}
        <label class="field">
          <span>{t('auth.authentication_code_or_recovery_code')}</span>
          <!-- svelte-ignore a11y_autofocus -->
          <input
            class="input code"
            name="code"
            inputmode="text"
            autocomplete="one-time-code"
            bind:value={code}
            autofocus={!challenge}
            required
          />
        </label>
      {/if}
      {#if error}<p class="error" role="alert" tabindex="-1" use:focusOnShow>{error}</p>{/if}
      {#if !passwordHidden || needsCode || ticket}
        <button class="btn btn-primary wide" type="submit" disabled={busy}
          >{needsCode || ticket ? t('auth.verify') : t('auth.log_in')}</button
        >
      {/if}
      {#if !needsCode}
        <!-- A real second button, as large as the first: guests are people too (#513). -->
        <button class="btn wide" type="button" onclick={() => ((asGuest = !asGuest), (error = ''))}
          >{asGuest ? t('auth.sign_in_as_owner') : t('auth.sign_in_as_guest')}</button
        >
      {/if}
      {#if needsCode}
        <button
          class="btn btn-icon"
          type="button"
          onclick={() => {
            needsCode = false;
            code = error = '';
          }}>{t('auth.back')}</button
        >
      {/if}
      <!-- Members and guests only: the admin resets on the server. -->
      {#if asGuest && auth.passwordReset && !needsCode}
        <button class="btn btn-icon" type="button" onclick={() => ((forgot = true), (error = ''))}
          >{t('auth.forgot_password')}</button
        >
      {/if}
    </form>
  {/if}
  {#if poweredBy}<p class="powered muted">{t('auth.powered_by')}</p>{/if}
</AuthCard>

<style>
  .powered {
    margin: var(--space-3) 0 0;
    text-align: center;
    font-size: 0.8em;
  }
  .hidden {
    display: none;
  }
  form {
    display: grid;
    gap: var(--space-2);
  }
  .wide {
    width: 100%;
    justify-content: center;
    min-height: 2.6rem;
  }
  .divider {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin: var(--space-3) 0;
    font-size: 0.85em;
    color: var(--color-text-muted);
  }
  .divider::before,
  .divider::after {
    content: '';
    flex: 1;
    border-top: 1px solid var(--color-border);
  }
  .error:focus {
    outline: none;
  }
  .code {
    font-family: var(--font-mono);
    letter-spacing: 0.15em;
    font-size: 1.2rem;
  }
</style>
