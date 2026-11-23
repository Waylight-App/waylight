<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '../../../../shared/i18n';
  import Icon from '../../../../shared/ui/icons/Icon.svelte';
  import { doneProgress, type LaterSetup } from '../../domain/wizard';
  import type { StepProps } from './wizardContext';

  let { ctx, session, actions }: StepProps = $props();

  const progress = $derived(doneProgress(session.created, session.place !== undefined));
  const total = $derived(session.created.length);
  const percent = $derived(total === 0 ? 0 : Math.round((progress.ready.length / total) * 100));
  const readyNames = $derived(progress.ready.map((kind) => ctx.widgetName(kind)));

  // What each open widget needs and where it is explained; the list shows the
  // widget's own description until the help has arrived.
  let help = $state<Record<string, { needs: string; docsUrl: string }>>({});
  onMount(() => {
    for (const { kind } of progress.open) {
      void ctx.setupHelp(kind).then((h) => {
        if (h) help[kind] = h;
      });
    }
  });

  /** Ends the wizard and leads through the given widgets on the dashboard. */
  function setUp(kinds: string[]) {
    ctx.startSetup(kinds);
    actions.next();
  }

  const label = (action: LaterSetup) =>
    t(action === 'connect' ? 'wizard.done.action_connect' : 'wizard.done.action_set_up');
</script>

{#if total === 0}
  <div class="wz-badge">
    <svg
      width="40"
      height="40"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2.2"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"><path d="M5 12l5 5 9-10" /></svg
    >
  </div>
  <h1 class="wz-title wz-title-lg" tabindex="-1">{t('wizard.done.title')}</h1>
  <p class="wz-lead">{t('wizard.done.body')}</p>
  <button class="wz-btn wz-btn-lg" type="button" onclick={actions.next}>{t('wizard.done.finish')}</button>
{:else}
  <div class="wz-done">
    <div class="wz-done-left">
      <div
        class="wz-ring"
        style:--wz-percent="{percent}%"
        role="img"
        aria-label={t('wizard.done.ring', { ready: progress.ready.length, total })}
      >
        <div>
          <b>{progress.ready.length}/{total}</b>
          <small>{t('wizard.done.ready_short')}</small>
        </div>
      </div>
      <h1 class="wz-title wz-title-lg" tabindex="-1">
        {progress.open.length > 0 ? t('wizard.done.title_open') : t('wizard.done.title')}
      </h1>
      <p class="wz-lead">
        {progress.open.length > 0
          ? t('wizard.done.lead_open', { ready: progress.ready.length, total })
          : t('wizard.done.body')}
      </p>
      <div class="wz-done-actions">
        {#if progress.open.length > 0}
          <button class="wz-btn" type="button" onclick={() => setUp(progress.open.map((o) => o.kind))}
            >{t('wizard.done.set_up_now')}</button
          >
          <button class="wz-btn wz-btn-line" type="button" onclick={actions.next}
            >{t('wizard.done.finish')}</button
          >
        {:else}
          <button class="wz-btn" type="button" onclick={actions.next}>{t('wizard.done.finish')}</button>
        {/if}
      </div>
      {#if progress.open.length > 0}<p class="wz-done-note">{t('wizard.done.optional')}</p>{/if}
    </div>
    <div class="wz-done-right">
      {#if progress.open.length > 0}
        <h2 class="wz-eyebrow">{t('wizard.done.open', { n: progress.open.length })}</h2>
        <ul class="wz-todo">
          {#each progress.open as item (item.kind)}
            {@const icon = ctx.widgetIcon(item.kind)}
            <li>
              <span class="wz-todo-icon" aria-hidden="true"
                >{#if icon}<Icon name={icon} size={22} />{/if}</span
              >
              <span class="wz-todo-text">
                <b>{ctx.widgetName(item.kind)}</b>
                <span>{help[item.kind]?.needs ?? ''}</span>
              </span>
              {#if help[item.kind]}
                <a href={help[item.kind]?.docsUrl} target="_blank" rel="noopener noreferrer"
                  >{t('wizard.done.where')}</a
                >
              {/if}
              <button
                class="wz-mini"
                type="button"
                aria-label={t('wizard.done.action_for', {
                  action: label(item.action),
                  name: ctx.widgetName(item.kind),
                })}
                onclick={() => setUp([item.kind])}>{label(item.action)}</button
              >
            </li>
          {/each}
        </ul>
      {/if}
      {#if readyNames.length > 0}
        <h2 class="wz-eyebrow">{t('wizard.done.ready')}</h2>
        <p class="wz-ready">
          <svg
            width="22"
            height="22"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.4"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"><path d="M5 12l5 5 9-10" /></svg
          >
          <span class="wz-ready-names">{readyNames.join(' · ')}</span>
        </p>
      {/if}
    </div>
  </div>
{/if}
