<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { intlLocale, t } from '../../../shared/i18n';
  import { linkResult, type LinkedSignIn, type LinkResult } from '../domain/identity';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /**
   * Settings → Account: link a sign-in at the provider, such as a personal
   * Microsoft account, and sign in with it in one click (#527).
   */
  interface Props {
    auth: AuthUseCases;
  }
  let { auth }: Props = $props();

  const provider = $derived(auth.sso.buttonLabel || t('auth.linked_provider'));
  let links = $state<LinkedSignIn[]>([]);
  let error = $state('');
  let busy = $state(false);

  // The server sends people back here with the result of linking.
  let result = $state<LinkResult | null>(null);
  const params = new URLSearchParams(location.search);
  result = linkResult(params);
  if (params.has('link')) {
    params.delete('link');
    const q = params.toString();
    history.replaceState(history.state, '', location.pathname + (q ? `?${q}` : '') + location.hash);
  }
  const resultKey = {
    ok: 'auth.linked_ok',
    taken: 'auth.linked_taken',
    licence: 'auth.linked_licence',
  } as const;

  const load = () =>
    auth
      .linkedSignIns()
      .then((l) => (links = l))
      .catch((err) => (error = errorMessage(err)));

  $effect(() => {
    load();
  });

  async function connect() {
    busy = true;
    error = '';
    try {
      location.assign(await auth.beginLinkSignIn());
    } catch (err) {
      error = errorMessage(err);
      busy = false;
    }
  }

  async function disconnect(l: LinkedSignIn) {
    if (!confirm(t('auth.linked_disconnect_confirm', { name: l.email || provider }))) return;
    try {
      await auth.unlinkSignIn(l.subject);
      result = null;
      await load();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<h3>{t('auth.linked_title', { name: provider })}</h3>
<p class="muted">{t('auth.linked_hint', { name: provider })}</p>
{#if result}
  <p class={result === 'ok' ? 'ok' : 'error'} role="status">{t(resultKey[result], { name: provider })}</p>
{/if}
{#if links.length}
  <ul class="links">
    {#each links as l (l.subject)}
      <li>
        <span
          >{l.email || provider}
          <span class="muted small"
            >{t('auth.linked_since', { date: l.linkedAt.toLocaleDateString(intlLocale()) })}</span
          ></span
        >
        <button class="btn" type="button" onclick={() => disconnect(l)}>{t('auth.linked_disconnect')}</button>
      </li>
    {/each}
  </ul>
{:else}
  <button class="btn" type="button" disabled={busy} onclick={connect}
    >{t('auth.linked_connect', { name: provider })}</button
  >
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}

<style>
  .links {
    list-style: none;
    padding: 0;
    margin: var(--space-2) 0;
    display: grid;
    gap: var(--space-2);
  }
  .links li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-2);
  }
  .small {
    font-size: 0.85em;
  }
  .ok {
    color: var(--color-success);
  }
</style>
