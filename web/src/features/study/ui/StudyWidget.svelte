<script lang="ts">
  import Meter from '../../../shared/ui/Meter.svelte';
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { formatEcts, formatGrade, type Progress } from '../domain/study';
  import type { StudyUseCases } from '../usecase/studyUseCases';

  /** Props of the study progress widget (#57). */
  interface Props {
    study: StudyUseCases;
  }
  let { study }: Props = $props();
  const poll = usePolling();

  // Local data that only changes in the settings; a look every few minutes keeps other tabs in sync.
  const POLL_MS = 5 * 60_000;
  let progress = $state<Progress | null>(null);
  let error = $state('');

  $effect(() =>
    poll(
      () =>
        study
          .progress()
          .then((p) => ((progress = p), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      POLL_MS,
    ),
  );

  const ects = (v: number) => formatEcts(v, intlLocale());
</script>

{#if error && !progress}
  <p class="error">{error}</p>
{:else if progress}
  {#if progress.modules === 0}
    <p class="muted">{t('study.set_up')}</p>
  {:else}
    {@const p = progress}
    <!-- Passed ECTS is the one figure of this widget, so its bar takes the accent (#814). -->
    <Meter
      label={t('study.of_ects', { passed: ects(p.passedEcts), total: ects(p.totalEcts) })}
      percent={p.percent}
      detail={t('study.percent', { n: Math.round(p.percent) })}
      hot
      valueText={t('study.of_ects', { passed: ects(p.passedEcts), total: ects(p.totalEcts) })}
      extra={(p.inProgressEcts / p.totalEcts) * 100}
    />
    <p class="muted small">{t('study.passed_modules', { n: p.passedModules, total: p.modules })}</p>
    {#if p.average !== null}
      <p class="average" title={t('study.average_hint', { ects: ects(p.gradedEcts) })}>
        {t('study.average', { grade: formatGrade(p.average, intlLocale()) })}
      </p>
    {/if}
    {#if p.inProgress.length}
      <h4>{t('study.in_progress')}</h4>
      <ul>
        {#each p.inProgress as m (m.id)}
          <ListRow nowrap>
            {m.name}
            {#snippet value()}{ects(m.ects)} ECTS{/snippet}
          </ListRow>
        {/each}
      </ul>
    {/if}
  {/if}
{/if}

<style>
  p {
    margin: 0;
  }
  .average {
    margin-top: var(--space-1);
    font-size: var(--fs-sm);
    font-weight: 600;
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
