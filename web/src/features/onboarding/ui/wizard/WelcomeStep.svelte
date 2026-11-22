<script lang="ts">
  import { t } from '../../../../shared/i18n';
  import type { StepProps } from './wizardContext';

  let { actions }: StepProps = $props();
</script>

<h1 class="wz-title wz-title-xl" tabindex="-1">{t('wizard.welcome.title')}</h1>
<p class="wz-lead">{t('wizard.welcome.body')}</p>
<div class="wz-cta">
  <button class="wz-btn" type="button" onclick={actions.next}>{t('wizard.welcome.start')}</button>
  <button class="wz-link" type="button" onclick={actions.skipAll}>{t('wizard.welcome.later')}</button>
</div>
