<script lang="ts">
  import { loadLanguage, t } from '../shared/i18n';
  // Imported from the domain module, not the feature's index: the index also
  // re-exports every auth settings component (and the QR code library behind
  // two-factor setup), which would otherwise land in the initial bundle.
  import { accountLinkPath, type AuthView } from '../features/auth/domain/auth';
  import { errorMessage } from '../shared/api-client/http';
  import { canStartOffline, clearOfflineData, rememberSignedIn } from '../shared/pwa';
  import { createContainer, type Container } from './container';
  import { branding, loadBranding } from './branding/branding.svelte';
  import { loadDefaultTheme, showDefaultTheme } from './branding/defaultTheme';
  import { lastLook } from './branding/lastLook';
  import type { DefaultThemeDTO } from '../shared/api-client/generated/appearance';
  import { applyTheme } from '../shared/theme/theme';
  // The dashboard shell is fetched right away, in parallel with the session
  // check, but kept out of the initial bundle (#116); sign-in pages load
  // only when needed.
  const shell = import('./Shell.svelte');

  // The browser extension's local mode passes its own container (ADR 0019);
  // the web app builds the one that talks to the server.
  let { localContainer }: { localContainer?: Container } = $props();

  let view = $state<AuthView | 'loading'>('loading');
  // An invitation or reset link (#296): like a kiosk link, the token is in
  // the fragment; it is kept in memory and removed from the address bar, so
  // it does not stay in the history.
  let linkToken = $state('');
  if (location.pathname === accountLinkPath && location.hash.length > 1) {
    linkToken = decodeURIComponent(location.hash.slice(1));
    history.replaceState(null, '', '/');
  }
  let loadError = $state('');
  let defaultTheme = $state<DefaultThemeDTO>({ id: '', css: '', scheme: '' });
  // The default theme's style only while nobody is signed in: the shell
  // brings the account's appearance, which already includes it if the
  // account has no appearance of its own.
  $effect(() => (view === 'app' ? undefined : showDefaultTheme(defaultTheme)));

  // The container is built once; a changed prop would not rebuild it.
  // svelte-ignore state_referenced_locally
  const container = localContainer ?? createContainer(() => (view = 'login'));

  async function refresh() {
    try {
      // The browser language's catalogue first, so the page never flashes English.
      await loadLanguage().catch(() => undefined);
      // The organisation's name and colours (#305) show from the sign-in page on.
      const brand = await loadBranding(container.http);
      // Before sign-in the account is not known yet. The look of the account
      // that last signed in on this browser comes first (#526), else the
      // instance's default theme (#507) with the organisation's scheme. The
      // shell applies the person's own preference after sign-in.
      const look = container.localMode ? null : lastLook();
      if (look) defaultTheme = { id: '', css: look.css, scheme: '' };
      else if (!container.localMode) defaultTheme = await loadDefaultTheme(container.http);
      const scheme = look?.theme || brand.defaultTheme || defaultTheme.scheme;
      if (scheme) applyTheme(scheme);
      // A kiosk link (#50): the token is in the fragment, so it never reaches
      // the server log; it signs this browser in as the display.
      if (location.pathname === '/kiosk' && location.hash.length > 1) {
        const token = decodeURIComponent(location.hash.slice(1));
        history.replaceState(null, '', '/');
        await container.auth.kioskLogin(token);
      }
      view = await container.auth.currentView();
      loadError = '';
      rememberSignedIn(view === 'app');
      if (view !== 'app') await clearOfflineData();
    } catch (err) {
      // Offline after an earlier session: show the last loaded data read-only.
      if (canStartOffline()) view = 'app';
      else loadError = errorMessage(err);
    }
  }

  refresh();
</script>

{#if loadError}
  <p class="error" role="alert">{t('app.could_not_reach_the_server', { loadError: loadError })}</p>
{:else if linkToken && view !== 'loading'}
  {#await import('../features/auth/ui/AccountLinkPage.svelte') then { default: AccountLinkPage }}
    <AccountLinkPage
      auth={container.auth}
      token={linkToken}
      ondone={() => {
        linkToken = '';
        refresh();
      }}
    />
  {/await}
{:else if view === 'setup'}
  {#await import('../features/auth/ui/SetupPage.svelte') then { default: SetupPage }}
    <SetupPage auth={container.auth} ondone={refresh} />
  {/await}
{:else if view === 'login'}
  {#await import('../features/auth/ui/LoginPage.svelte') then { default: LoginPage }}
    <LoginPage
      auth={container.auth}
      ondone={refresh}
      brandName={branding().name}
      poweredBy={branding().custom && !branding().hidePoweredBy}
      logo={branding().logo}
      logoDark={branding().logoDark}
    />
  {/await}
{:else if view === 'app' && location.pathname === '/capture'}
  <!-- Quick capture (#66): a page of its own, without the dashboard. -->
  {#await import('./CapturePage.svelte') then { default: CapturePage }}
    <CapturePage {container} />
  {/await}
{:else if view === 'app'}
  {#if container.auth.demo}
    <p class="demo-banner" role="note">{t('app.demo_banner')}</p>
  {/if}
  {#await shell then { default: Shell }}
    <Shell {container} onlogout={refresh} />
  {:catch err}
    <p class="error" role="alert">{t('app.could_not_reach_the_server', { loadError: errorMessage(err) })}</p>
  {/await}
{/if}

<style>
  .demo-banner {
    margin: 0;
    padding: var(--space-1) var(--space-3);
    text-align: center;
    font-size: 0.9em;
    background: var(--color-accent);
    color: var(--color-accent-contrast);
  }
</style>
