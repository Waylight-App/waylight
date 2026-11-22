<script lang="ts">
  import { tick } from 'svelte';
  import { t } from '../../../../shared/i18n';
  import { errorMessage } from '../../../../shared/api-client/http';
  import type { WizardStepId } from '../../domain/wizard';
  import { stepRegistry } from './steps';
  import type { StepActions, WizardContext } from './wizardContext';
  import { StepControl, WizardSession } from './wizardState.svelte';
  import './wizard.css';

  /**
   * The onboarding wizard (#739): a full-screen view before the dashboard,
   * one step per screen with a progress bar, back and skip. It lays the
   * steps out; what each step asks lives in its own component.
   */
  interface Props {
    steps: WizardStepId[];
    /** Index of the step to open at (a resumed wizard). */
    start: number;
    ctx: WizardContext;
    /** Remembers the step so a reload continues there. */
    onstep: (step: WizardStepId) => void;
    /** The wizard is over, finished or skipped. */
    onfinish: (session: WizardSession) => void;
  }
  let { steps, start, ctx, onstep, onfinish }: Props = $props();

  // svelte-ignore state_referenced_locally
  let index = $state(start);
  const step = $derived(steps[index] ?? 'done');
  const extraId = $derived(step.startsWith('extra:') ? step.slice('extra:'.length) : undefined);
  const registered = $derived(
    extraId === undefined ? stepRegistry[step as keyof typeof stepRegistry] : undefined,
  );
  const glow = $derived(registered?.glow ?? 'top');
  const centered = $derived(registered?.centered ?? true);
  const footer = $derived(registered?.footer ?? 'actions');
  const Step = $derived(registered?.component);
  const extraTitle = $derived(ctx.extraSteps.find((s) => s.id === extraId)?.title ?? '');

  const session = new WizardSession();
  let ctl = $state.raw(new StepControl());
  let busy = $state(false);
  let error = $state('');
  let content: HTMLElement | undefined = $state();

  const progress = $derived(t('wizard.progress', { n: index + 1, total: steps.length }));

  function go(next: number) {
    error = '';
    if (next >= steps.length) {
      ctx.preview(null);
      onfinish(session);
      return;
    }
    if (next < 0) return;
    index = next;
    ctl = new StepControl();
    onstep(steps[next] as WizardStepId);
  }

  /** Runs the step's save, then moves on; a failure stays on the step. */
  async function next() {
    if (busy) return;
    busy = true;
    error = '';
    try {
      await ctl.save?.();
      go(index + 1);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  function skip() {
    ctl.skip?.();
    go(index + 1);
  }

  const actions: StepActions = {
    next: () => void next(),
    skipAll: () => {
      ctx.preview(null);
      onfinish(session);
    },
  };

  // A new screen starts at its heading, for keyboard and screen reader users.
  $effect(() => {
    void index;
    void tick().then(() => content?.querySelector<HTMLElement>('h1')?.focus());
  });
</script>

<div class="wz" class:wz-glow-middle={glow === 'middle'} role="region" aria-label={t('wizard.label')}>
  <header class="wz-top">
    <div class="wz-logo">
      <svg
        width="30"
        height="30"
        viewBox="0 0 32 32"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
        ><path d="M11 22c0-3-3-4.5-3-8a8 8 0 0 1 16 0c0 3.5-3 5-3 8z" /><path d="M12 26h8M13 29.5h6" /></svg
      >Waylight
    </div>
    <div
      class="wz-dots"
      role="progressbar"
      aria-valuemin={1}
      aria-valuemax={steps.length}
      aria-valuenow={index + 1}
      aria-valuetext={progress}
      aria-label={t('wizard.progress_label')}
    >
      {#each steps as id, i (id)}<i class:on={i <= index}></i>{/each}
    </div>
  </header>

  <main class="wz-main" class:wz-main-center={centered} bind:this={content}>
    {#if Step}
      <Step {ctx} {session} {ctl} {actions} />
    {:else if extraId !== undefined && ctx.extra}
      <h1 class="wz-title" tabindex="-1">{extraTitle}</h1>
      <div class="wz-form">
        {@render ctx.extra(extraId, (save) => (ctl.save = save))}
      </div>
    {/if}
    {#if error}<p class="wz-error" role="alert">{error}</p>{/if}
  </main>

  {#if footer === 'hint'}
    <footer class="wz-foot wz-foot-hint">
      {#if registered?.hint}{t(registered.hint)}{/if}
    </footer>
  {:else}
    <footer class="wz-foot">
      {#if index > 0}
        <button class="wz-link" type="button" disabled={busy} onclick={() => go(index - 1)}
          >{t('wizard.back')}</button
        >
      {:else}<span></span>{/if}
      <div class="wz-foot-end">
        {#if ctl.status}<span class="wz-status" role="status">{ctl.status}</span>{/if}
        <button class="wz-link" type="button" disabled={busy} onclick={skip}>{t('wizard.skip')}</button>
        <button class="wz-btn" type="button" disabled={busy} onclick={() => void next()}
          >{t('wizard.next')}</button
        >
      </div>
    </footer>
  {/if}
</div>
