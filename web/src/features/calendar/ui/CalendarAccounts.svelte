<script lang="ts">
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import {
    CALENDAR_PROVIDERS,
    type CalendarAccount,
    type CalendarProvider,
    PROVIDER_CONSOLE,
  } from '../domain/accounts';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';

  /** Settings → Calendar → Google and Microsoft accounts (#282). */
  interface Props {
    calendar: CalendarUseCases;
  }
  let { calendar }: Props = $props();

  let accounts = $state<CalendarAccount[]>([]);
  let redirect = $state('');
  let loaded = $state(false);
  let provider = $state<CalendarProvider>('google');
  let clientId = $state('');
  let clientSecret = $state('');
  let error = $state('');
  let message = $state('');
  let busy = $state(false);
  let stopped = false;

  const providerName = (p: CalendarProvider) => (p === 'microsoft' ? 'Microsoft' : 'Google');

  async function load() {
    const res = await calendar.accounts();
    accounts = res.accounts;
    redirect = res.redirectUrl;
    loaded = true;
  }

  $effect(() => {
    load().catch((err) => ((error = errorMessage(err)), (loaded = true)));
    return () => (stopped = true);
  });

  /** Opens the provider's sign-in in a new window and waits for it. */
  function signIn(e: SubmitEvent) {
    e.preventDefault();
    // Opened synchronously in the submit, so pop-up blockers allow it.
    const win = window.open('about:blank', 'trisse-calendar-oauth', 'width=520,height=700');
    if (!win) {
      error = t('calendar.accounts.popup_blocked');
      return;
    }
    win.opener = null; // the provider's page gets no handle on this one
    error = message = '';
    busy = true;
    void (async () => {
      try {
        const start = await calendar.startSignIn(provider, clientId, clientSecret);
        win.location.href = start.url;
        message = t('calendar.accounts.signing_in');
        const status = await calendar.waitForSignIn(start.state, () => stopped);
        message = '';
        if (status === 'failed') error = t('calendar.accounts.sign_in_failed');
        if (status === 'ready') {
          clientSecret = '';
          message = t('calendar.accounts.connected');
          await load();
        }
      } catch (err) {
        win.close();
        error = errorMessage(err);
      } finally {
        busy = false;
      }
    })();
  }

  async function toggle(a: CalendarAccount, id: string, on: boolean) {
    const enabled = a.calendars.filter((c) => (c.id === id ? on : c.enabled)).map((c) => c.id);
    try {
      const updated = await calendar.setCalendars(a.id, enabled);
      accounts = accounts.map((x) => (x.id === a.id ? updated : x));
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function disconnect(a: CalendarAccount) {
    if (!confirm(t('calendar.accounts.disconnect_confirm', { name: a.name }))) return;
    try {
      await calendar.disconnect(a.id);
      accounts = accounts.filter((x) => x.id !== a.id);
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<h3>{t('calendar.accounts.title')}</h3>
<p class="muted">{t('calendar.accounts.intro')}</p>

{#each accounts as a (a.id)}
  <section class="account" aria-label={`${providerName(a.provider)}: ${a.name}`}>
    <header>
      <strong>{providerName(a.provider)}</strong> <span class="muted">{a.name}</span>
      <button class="btn" type="button" onclick={() => disconnect(a)}
        >{t('calendar.accounts.disconnect')}</button
      >
    </header>
    <ul>
      {#each a.calendars as c (c.id)}
        <li>
          <label class="check">
            <input
              type="checkbox"
              checked={c.enabled}
              onchange={(e) => toggle(a, c.id, e.currentTarget.checked)}
            />
            {#if c.color}<span class="swatch" style:background={c.color} aria-hidden="true"></span>{/if}
            {c.name}
            {#if !c.writable}<span class="muted small">{t('calendar.accounts.read_only')}</span>{/if}
          </label>
        </li>
      {/each}
    </ul>
  </section>
{/each}

{#if loaded && redirect}
  <form onsubmit={signIn}>
    <label class="field"
      ><span>{t('calendar.accounts.provider')}</span><select class="input" bind:value={provider}>
        {#each CALENDAR_PROVIDERS as p (p)}<option value={p}>{providerName(p)}</option>{/each}
      </select></label
    >
    <p class="muted small">
      {t('calendar.accounts.register', { service: providerName(provider) })}
      <a href={PROVIDER_CONSOLE[provider]} target="_blank" rel="noopener noreferrer"
        >{t('calendar.accounts.console', { service: providerName(provider) })}</a
      >
    </p>
    <p><code class="redirect">{redirect}</code></p>
    <label class="field"
      ><span>{t('calendar.accounts.client_id')}</span><input
        class="input"
        bind:value={clientId}
        required
        maxlength="200"
        autocomplete="off"
      /></label
    >
    <label class="field"
      ><span>{t('calendar.accounts.client_secret')}</span><input
        class="input"
        type="password"
        bind:value={clientSecret}
        maxlength="500"
        autocomplete="off"
      /></label
    >
    <button class="btn btn-primary" type="submit" disabled={busy}
      >{t('calendar.accounts.sign_in', { service: providerName(provider) })}</button
    >
  </form>
{:else if loaded}
  <p class="muted">{t('calendar.accounts.unavailable')}</p>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  h3 {
    margin: var(--space-3) 0 var(--space-1);
    font-size: 1em;
  }
  .account {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small, 6px);
    padding: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .account header {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
  }
  .account header button {
    margin-left: auto;
  }
  ul {
    list-style: none;
    margin: var(--space-1) 0 0;
    padding: 0;
    display: grid;
    gap: var(--space-1);
  }
  .check {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  .swatch {
    width: 0.8rem;
    height: 0.8rem;
    border-radius: 50%;
  }
  form {
    display: grid;
    gap: var(--space-2);
  }
  .redirect {
    word-break: break-all;
  }
  .small {
    font-size: 0.85em;
  }
</style>
