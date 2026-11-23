<script lang="ts">
  import { untrack } from 'svelte';
  import { errorMessage } from '../../../shared/api-client/http';
  import { t } from '../../../shared/i18n';
  import Modal from '../../../shared/ui/Modal.svelte';
  import { copyToClipboard } from '../../../shared/ui/clipboard';
  import type { MicrosoftFlow } from '../domain/identity';
  import type { AuthUseCases } from '../usecase/authUseCases';

  /**
   * Sign in or link with a Microsoft account by code (#535). Microsoft's page
   * opens in a window, this dialog shows the code, and the server is asked at
   * Microsoft's interval until the code is confirmed. Nothing is redirected
   * back to this server, so it works on any address.
   */
  interface Props {
    auth: AuthUseCases;
    /** Link to the signed-in account, or sign in. */
    mode: 'link' | 'signin';
    /** The account is linked, or signed in (ticket: the local code is needed). */
    ondone: (result: { linked?: boolean; ticket?: string }) => void;
    onclose: () => void;
  }
  let { auth, mode, ondone, onclose }: Props = $props();

  let flow = $state<MicrosoftFlow | null>(null);
  let error = $state('');
  let copied = $state(false);
  let popup: Window | null = null;
  let stopped = false;

  function openMicrosoft() {
    if (!flow) return;
    // A window, not a tab, so this dialog stays visible next to it.
    popup = window.open(flow.verificationUri, 'waylight-microsoft', 'popup,width=520,height=700');
  }

  async function copy() {
    if (flow) copied = await copyToClipboard(flow.userCode);
  }

  async function poll(f: MicrosoftFlow) {
    while (!stopped) {
      await new Promise((r) => setTimeout(r, f.interval * 1000));
      if (stopped) return;
      try {
        const res = await auth.pollMicrosoft(f.flow);
        if ('pending' in res) continue;
        stopped = true;
        popup?.close();
        ondone('linked' in res ? { linked: true } : 'ticket' in res ? { ticket: res.ticket } : {});
        return;
      } catch (err) {
        stopped = true;
        popup?.close();
        error = errorMessage(err);
        return;
      }
    }
  }

  async function start() {
    error = '';
    stopped = false;
    try {
      const f = await auth.startMicrosoft(mode === 'link');
      flow = f;
      // The code is in the clipboard before the window asks for it.
      copied = await copyToClipboard(f.userCode);
      openMicrosoft();
      void poll(f);
    } catch (err) {
      error = errorMessage(err);
    }
  }

  $effect(() => {
    untrack(() => void start());
    return () => {
      stopped = true;
      popup?.close();
    };
  });
</script>

<Modal title={t('auth.ms_title')} {onclose}>
  {#if flow && !error}
    <p>{t('auth.ms_enter_code')}</p>
    <p class="code" aria-live="polite">{flow.userCode}</p>
    <div class="row">
      <button class="btn" type="button" onclick={copy}
        >{copied ? t('auth.ms_copied') : t('auth.ms_copy')}</button
      >
      <button class="btn" type="button" onclick={openMicrosoft}>{t('auth.ms_open_again')}</button>
    </div>
    <p class="muted small" role="status">{t('auth.ms_waiting')}</p>
  {:else if error}
    <p class="error" role="alert">{error}</p>
    <button class="btn btn-primary" type="button" onclick={start}>{t('auth.ms_try_again')}</button>
  {:else}
    <p class="muted" role="status">{t('auth.ms_starting')}</p>
  {/if}
</Modal>

<style>
  .code {
    margin: var(--space-3) 0;
    font-family: var(--font-mono);
    font-size: var(--fs-lg);
    letter-spacing: 0.15em;
    text-align: center;
  }
  .row {
    display: flex;
    gap: var(--space-2);
    justify-content: center;
    flex-wrap: wrap;
  }
  .small {
    font-size: var(--fs-xs);
    text-align: center;
  }
</style>
