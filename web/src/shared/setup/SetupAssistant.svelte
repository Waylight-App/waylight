<script lang="ts">
  import { onDestroy, tick } from 'svelte';
  import { t } from '../i18n';
  import { errorMessage } from '../api-client/http';
  import {
    addressOf,
    type FoundService,
    initialValues,
    type SetupDescriptor,
    type SetupField,
    type SetupPreview,
    type SetupStage,
    signInAddress,
    stageComplete,
  } from './setup';
  import { followSignIn, useSignIn } from './signin';

  /**
   * Props of the setup assistant shown in place of a widget that is not set
   * up (#588): address, credential with the service's own steps, then a
   * check that shows real data before the widget takes over.
   */
  interface Props {
    descriptor: SetupDescriptor;
    /** Called after Done; the widget loads its data again. */
    ondone: () => void;
  }
  let { descriptor: d, ondone }: Props = $props();

  type Stage = SetupStage | 'done';
  let stage = $state<Stage>('address');
  // svelte-ignore state_referenced_locally
  let values = $state<Record<string, string>>(initialValues(d));
  let error = $state('');
  let checking = $state(false);
  let preview = $state<SetupPreview | null>(null);
  let found = $state<FoundService[]>([]);
  let heading = $state<HTMLElement>();

  // Signing in with the service (#654) comes first where the descriptor
  // offers it and the shell provides the server side; the pasted key stays
  // one click away and takes over when the sign-in cannot start.
  const signIn = useSignIn();
  let mode = $state<'signin' | 'manual'>('signin');
  let signInStatus = $state('');
  let loginLink = $state('');
  let running: AbortController | null = null;
  const signInFrom = $derived(signIn ? signInAddress(d, values) : null);
  const signingIn = $derived(stage === 'credential' && mode === 'signin' && !!signInFrom);
  const signInLabel = $derived(d.signIn?.label ?? d.service);

  const index = $derived(stage === 'address' ? 1 : stage === 'credential' ? 2 : 3);
  const fieldsOf = (s: SetupStage) => d.fields.filter((f) => f.stage === s);
  const link = $derived(
    stage === 'credential' ? (d.credentialPage?.(addressOf(d, values), values) ?? null) : null,
  );

  // Services found in the home network are offered first (#590); the
  // assistant works the same without them.
  $effect(() => {
    d.discover?.()
      .then((list) => (found = list))
      .catch(() => (found = []));
  });

  function stopSignIn() {
    running?.abort();
    running = null;
    signInStatus = '';
    loginLink = '';
  }
  onDestroy(stopSignIn);

  function useMode(next: 'signin' | 'manual') {
    stopSignIn();
    error = '';
    mode = next;
  }

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
          loginLink = url;
          signInStatus = t('setup.signin_waiting');
          if (win && !win.closed) win.location.href = url;
        },
        controller.signal,
      );
      signInStatus = t('setup.signin_checking');
      preview = await d.signIn.check({ ...values }, id);
      running = null;
      await go('done');
    } catch (err) {
      if (controller.signal.aborted) return;
      win?.close();
      running = null;
      signInStatus = '';
      error = errorMessage(err);
      // A sign-in that never started (no PUBLIC_URL, not the service)
      // leaves the pasted key as the way forward.
      if (!loginLink) mode = 'manual';
      loginLink = '';
    } finally {
      if (running === null || running === controller) checking = false;
    }
  }

  async function go(next: Stage) {
    if (next !== 'done') stopSignIn();
    error = '';
    stage = next;
    await tick();
    heading?.focus();
  }

  function useFound(s: FoundService) {
    const field = fieldsOf('address')[0];
    if (field) values[field.key] = s.url;
    void go('credential');
  }

  function next(e: SubmitEvent) {
    e.preventDefault();
    if (stageComplete(d, 'address', values)) void go('credential');
  }

  async function check(e: SubmitEvent) {
    e.preventDefault();
    if (!stageComplete(d, 'credential', values)) return;
    error = '';
    checking = true;
    try {
      preview = await d.check({ ...values });
      await go('done');
    } catch (err) {
      error = errorMessage(err);
    } finally {
      checking = false;
    }
  }
</script>

{#snippet input(f: SetupField)}
  <label class="field">
    <span>{t(f.label)}</span>
    {#if f.type === 'choice'}
      <select class="input" bind:value={values[f.key]} required>
        {#each f.options ?? [] as o (o.value)}<option value={o.value}>{o.label}</option>{/each}
      </select>
    {:else}
      <input
        class="input"
        type={f.type}
        placeholder={f.placeholder}
        autocomplete={f.autocomplete ?? (f.type === 'password' ? 'new-password' : 'off')}
        bind:value={values[f.key]}
        required={!f.optional}
      />
    {/if}
  </label>
{/snippet}

<section class="setup" aria-labelledby="setup-heading">
  <p class="muted step">{t('setup.step', { n: index, total: 3 })}</p>
  <h4 id="setup-heading" tabindex="-1" bind:this={heading}>
    {stage === 'done'
      ? t('setup.connected', { service: d.service })
      : t('setup.title', { service: d.service })}
  </h4>

  {#if stage === 'address'}
    {#if found.length}
      <ul class="found">
        {#each found as s (s.url)}
          <li>
            <span
              >{t('setup.found', { service: d.service, url: s.url })}<br /><small class="muted"
                >{s.hint}</small
              ></span
            >
            <button class="btn btn-primary" type="button" onclick={() => useFound(s)}
              >{t('setup.connect')}</button
            >
          </li>
        {/each}
      </ul>
    {/if}
    <form onsubmit={next}>
      {#each fieldsOf('address') as f (f.key)}{@render input(f)}{/each}
      <button class="btn btn-primary" type="submit">{t('setup.next')}</button>
    </form>
  {:else if signingIn}
    <p>{t('setup.signin_intro', { service: signInLabel })}</p>
    <div class="actions">
      <button class="btn btn-primary" type="button" onclick={startSignIn} disabled={checking}
        >{t('setup.signin', { service: signInLabel })}</button
      >
      {#if loginLink}
        <a class="btn" href={loginLink} target="_blank" rel="noopener noreferrer"
          >{t('setup.signin_reopen')}</a
        >
      {/if}
    </div>
    <p class="muted" role="status" aria-live="polite">{signInStatus}</p>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <div class="actions">
      <button class="btn" type="button" onclick={() => go('address')}>{t('setup.back')}</button>
      <button class="btn" type="button" onclick={() => useMode('manual')}>{t('setup.signin_manual')}</button>
    </div>
  {:else if stage === 'credential'}
    <ol class="steps">
      {#each d.steps as s (s)}<li>{t(s)}</li>{/each}
    </ol>
    {#if link}
      <a class="btn" href={link} target="_blank" rel="noopener noreferrer"
        >{t('setup.open', { service: d.service })}</a
      >
    {/if}
    <form onsubmit={check}>
      {#each fieldsOf('credential') as f (f.key)}{@render input(f)}{/each}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <div class="actions">
        <button class="btn" type="button" onclick={() => go('address')}>{t('setup.back')}</button>
        <button class="btn btn-primary" type="submit" disabled={checking}
          >{checking ? t('setup.checking') : t('setup.check')}</button
        >
        {#if signInFrom}
          <button class="btn" type="button" onclick={() => useMode('signin')}
            >{t('setup.signin_instead', { service: signInLabel })}</button
          >
        {/if}
      </div>
    </form>
  {:else if preview}
    <ul class="preview" role="status">
      {#each preview.lines as line (line)}<li>{line}</li>{/each}
    </ul>
    {#if preview.images?.length}
      <div class="images">
        {#each preview.images.slice(0, 3) as src (src)}<img {src} alt="" loading="lazy" />{/each}
      </div>
    {/if}
    <button class="btn btn-primary" type="button" onclick={ondone}>{t('setup.done')}</button>
  {/if}
</section>

<style>
  .setup {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .step {
    margin: 0;
    font-size: 0.8rem;
  }
  h4 {
    margin: 0;
  }
  h4:focus-visible {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  .steps {
    margin: 0;
    padding-left: 1.4em;
  }
  .steps li + li {
    margin-top: var(--space-1);
  }
  .found,
  .preview {
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
    gap: var(--space-2);
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-small);
  }
  .actions {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .images {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-1);
  }
  .images img {
    width: 100%;
    aspect-ratio: 1;
    object-fit: cover;
    border-radius: var(--radius-small);
  }
</style>
