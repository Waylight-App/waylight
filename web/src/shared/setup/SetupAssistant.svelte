<script lang="ts">
  import { onDestroy } from 'svelte';
  import { t } from '../i18n';
  import { errorMessage } from '../api-client/http';
  import Icon from '../ui/icons/Icon.svelte';
  import { qrDataUrl } from '../ui/qr';
  import { useWidgetIcon } from '../ui/widgetIcon';
  import SetupDialog from './SetupDialog.svelte';
  import SetupResult from './SetupResult.svelte';
  import SetupStep from './SetupStep.svelte';
  import {
    addressOf,
    fieldHint,
    type FoundService,
    initialValues,
    pasteTarget,
    type SetupDescriptor,
    type SetupField,
    type SetupPreview,
    signInAddress,
    stageComplete,
  } from './setup';
  import { followPaste, pasteLink, usePaste } from './paste';
  import { followSignIn, useSignIn } from './signin';

  /**
   * The guided setup of a widget (#588, #738), the same for every
   * integration: everything it needs on one screen as numbered steps, each
   * with a plain hint, one connection test that shows real data and one
   * primary action. Entering the key on the phone and signing in with the
   * service are offered next to the key's field.
   */
  interface Props {
    descriptor: SetupDescriptor;
    /** Called after Done, or when the dialog closes after a working connection. */
    ondone: () => void;
    /** Called when the setup is left without a working connection. */
    onclose: () => void;
  }
  let { descriptor: d, ondone, onclose }: Props = $props();
  const uid = $props.id();
  const widgetIcon = useWidgetIcon();

  // svelte-ignore state_referenced_locally
  let values = $state<Record<string, string>>(initialValues(d));
  let error = $state('');
  let checking = $state(false);
  let preview = $state<SetupPreview | null>(null);
  let found = $state<FoundService[]>([]);

  // Signing in with the service (#654) is offered where the descriptor has
  // it and the shell provides the server side; the key's field stays.
  const signIn = useSignIn();
  let signInStatus = $state('');
  let running: AbortController | null = null;
  const signInFrom = $derived(signIn ? signInAddress(d, values) : null);
  const signInLabel = $derived(d.signIn?.label ?? d.service);

  // Entering the key on the phone: the mobile app sends a key copied there;
  // it goes into the first secret field and is checked like a typed one.
  const paste = usePaste();
  const secret = $derived(pasteTarget(d));
  let pasteQr = $state('');
  let pasteStatus = $state('');
  let pasting: AbortController | null = null;

  const link = $derived(d.credentialPage?.(addressOf(d, values), values) ?? null);
  const complete = $derived(stageComplete(d, 'address', values) && stageComplete(d, 'credential', values));

  function stopPaste() {
    pasting?.abort();
    pasting = null;
    pasteQr = '';
    pasteStatus = '';
  }

  async function startPaste() {
    if (!paste || !secret) return;
    stopPaste();
    error = '';
    const controller = new AbortController();
    pasting = controller;
    try {
      const value = await followPaste(
        paste,
        (p) => {
          pasteQr = qrDataUrl(pasteLink(location.origin, p.code), 4);
          pasteStatus = t('setup.paste_waiting');
        },
        controller.signal,
      );
      values[secret.key] = value;
      pasting = null;
      pasteQr = '';
      pasteStatus = t('setup.paste_received');
    } catch (err) {
      if (controller.signal.aborted) return;
      pasting = null;
      pasteQr = '';
      pasteStatus = '';
      error = errorMessage(err);
    }
  }

  // Services found in the home network are offered first (#590); the
  // setup works the same without them.
  $effect(() => {
    d.discover?.()
      .then((list) => (found = list))
      .catch(() => (found = []));
  });

  function stopSignIn() {
    running?.abort();
    running = null;
    signInStatus = '';
  }
  onDestroy(() => {
    stopSignIn();
    stopPaste();
  });

  /**
   * Opens the sign-in window on the click itself, so no popup blocker
   * stops it, and points it at the service's page once the server named
   * it. The opener reference is cut before the service's page loads.
   */
  async function startSignIn() {
    if (!signIn || !d.signIn || !signInFrom) return;
    stopSignIn();
    error = '';
    const win = window.open('', '_blank');
    if (win) win.opener = null;
    const controller = new AbortController();
    running = controller;
    checking = true;
    try {
      const id = await followSignIn(
        signIn,
        d.signIn.service,
        signInFrom,
        (url) => {
          signInStatus = t('setup.signin_waiting');
          if (win && !win.closed) win.location.href = url;
        },
        controller.signal,
      );
      signInStatus = t('setup.signin_checking');
      preview = await d.signIn.check({ ...values }, id);
      signInStatus = '';
      running = null;
    } catch (err) {
      if (controller.signal.aborted) return;
      win?.close();
      running = null;
      signInStatus = '';
      error = errorMessage(err);
    } finally {
      if (running === null || running === controller) checking = false;
    }
  }

  async function check() {
    if (!complete) return;
    stopPaste();
    error = '';
    checking = true;
    try {
      preview = await d.check({ ...values });
    } catch (err) {
      error = errorMessage(err);
    } finally {
      checking = false;
    }
  }

  // The first press tests the connection and saves it; once it works the
  // same button is Done.
  const primary = () => (preview ? ondone() : void check());
  // A working connection is saved already, so leaving then still loads the widget.
  const leave = () => (preview ? ondone() : onclose());

  // A changed value makes the shown result obsolete.
  function changed() {
    preview = null;
    error = '';
  }

  const id = (f: SetupField) => `${uid}-${f.key}`;
</script>

{#snippet input(f: SetupField)}
  {#if f.type === 'choice'}
    <select id={id(f)} class="input" bind:value={values[f.key]} oninput={changed} required>
      {#each f.options ?? [] as o (o.value)}<option value={o.value}>{o.label}</option>{/each}
    </select>
  {:else}
    <input
      id={id(f)}
      class="input"
      type={f.type}
      placeholder={f.placeholder}
      autocomplete={f.autocomplete ?? (f.type === 'password' ? 'new-password' : 'off')}
      bind:value={values[f.key]}
      oninput={changed}
      required={!f.optional}
    />
  {/if}
{/snippet}

{#snippet actions(f: SetupField)}
  {#if f === secret && paste}
    <button
      class="btn ghost"
      type="button"
      aria-describedby={pasteQr ? `${uid}-qr` : undefined}
      onclick={pasteQr ? stopPaste : startPaste}
    >
      <Icon name="smartphone" size={18} />{pasteQr ? t('setup.paste_cancel') : t('setup.paste')}
    </button>
  {/if}
{/snippet}

<SetupDialog
  title={t('setup.title_connect', { service: d.service })}
  icon={widgetIcon}
  lead={d.intro ? t(d.intro) : t('setup.lead', { service: d.service })}
  primary={preview ? t('setup.done') : checking ? t('setup.checking') : t('setup.connect')}
  busy={checking}
  onprimary={primary}
  onclose={leave}
>
  {#if found.length && !addressOf(d, values)}
    <ul class="found">
      {#each found as s (s.url)}
        <li>
          <span
            >{t('setup.found', { service: d.service, url: s.url })}<br /><small class="muted">{s.hint}</small
            ></span
          >
          <button
            class="btn"
            type="button"
            onclick={() => {
              const f = d.fields.find((x) => x.stage === 'address' && x.type === 'url');
              if (f) values[f.key] = s.url;
            }}>{t('setup.use_found')}</button
          >
        </li>
      {/each}
    </ul>
  {/if}

  {#each d.fields as f, i (f.key)}
    <SetupStep n={i + 1} label={t(f.label)} for={id(f)}>
      {#snippet hint()}
        {#if f === secret}
          <ol>
            {#each d.steps as s (s)}<li>{t(s)}</li>{/each}
          </ol>
          {#if link}
            <a href={link} target="_blank" rel="noopener noreferrer"
              >{t('setup.open', { service: d.service })}</a
            >
          {/if}
        {:else}
          {fieldHint(f, d.service)}
        {/if}
      {/snippet}
      <div class="row">
        {@render input(f)}
        {@render actions(f)}
      </div>
      {#if f === secret}
        {#if pasteQr}
          <div class="qr-box" id="{uid}-qr">
            <img class="qr" src={pasteQr} alt={t('setup.paste_qr_alt')} />
            <p class="muted">{t('setup.paste_hint')}</p>
          </div>
        {/if}
        <p class="muted status" role="status" aria-live="polite">{pasteStatus}</p>
      {/if}
    </SetupStep>
  {/each}

  {#if signInFrom}
    <div class="signin">
      <button class="btn ghost" type="button" onclick={startSignIn} disabled={checking}
        >{t('setup.signin_instead', { service: signInLabel })}</button
      >
      <span class="muted" role="status" aria-live="polite">{signInStatus}</span>
    </div>
  {/if}

  {#if preview}
    <SetupResult ok lines={preview.lines} images={preview.images} />
  {:else if error}
    <SetupResult ok={false} lines={[error]} />
  {/if}
</SetupDialog>

<style>
  .row {
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }
  .row :global(.input) {
    flex: 1;
    min-width: 0;
  }
  .ghost {
    white-space: nowrap;
    color: var(--color-accent-text);
  }
  .status {
    margin: 0;
    min-height: 0;
  }
  .status:empty {
    display: none;
  }
  .signin {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-wrap: wrap;
  }
  .found {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .found li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
  }
  .qr-box {
    display: flex;
    align-items: center;
    gap: var(--space-4);
  }
  .qr-box p {
    margin: 0;
  }
  .qr {
    width: 10rem;
    height: 10rem;
    image-rendering: pixelated;
    background: #fff; /* scanners need a light quiet zone in dark themes */
  }
</style>
