<script lang="ts">
  import { t } from '../../../../shared/i18n';
  import { laterSetup } from '../../domain/wizard';
  import type { StepProps } from './wizardContext';

  let { ctx, session, actions }: StepProps = $props();

  const chips = $derived(
    session.created.map((kind) => ({
      kind,
      name: ctx.widgetName(kind),
      later: laterSetup(kind, session.place !== undefined),
    })),
  );
</script>

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
<p class="wz-lead">{chips.length > 0 ? t('wizard.done.body_widgets') : t('wizard.done.body')}</p>
{#if chips.length > 0}
  <ul class="wz-chips wz-chips-done" aria-label={t('wizard.done.list')}>
    {#each chips as chip (chip.kind)}
      <li class:todo={chip.later !== undefined}>
        {chip.later ? t(`wizard.done.${chip.later}`, { name: chip.name }) : chip.name}
      </li>
    {/each}
  </ul>
{/if}
<button class="wz-btn wz-btn-lg" type="button" onclick={actions.next}>{t('wizard.done.finish')}</button>
