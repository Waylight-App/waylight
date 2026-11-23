<script lang="ts">
  import EmptyState from '../../../shared/ui/EmptyState.svelte';
  import WasteSetup from './WasteSetup.svelte';
  import { useMaySetUp } from '../../../shared/setup/setup';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { dueLabel, isDueSoon, type WasteOverview } from '../domain/waste';
  import type { WasteUseCases } from '../usecase/wasteUseCases';

  /** Props of the waste widget. */
  interface Props {
    waste: WasteUseCases;
  }
  let { waste }: Props = $props();
  const poll = usePolling();
  const allowed = useMaySetUp(false);

  const POLL_MS = 30 * 60_000;
  let overview = $state<WasteOverview | null>(null);
  let error = $state('');
  let unconfigured = $state(false);
  let now = $state(new Date());
  let setupOpen = $state(false);
  let reload = $state(0);

  $effect(() => {
    void reload;
    const load = () =>
      waste
        .upcoming()
        .then((o) => ((overview = o), (now = new Date()), (error = ''), (unconfigured = false)))
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });

  const date = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { weekday: 'short', day: 'numeric', month: 'numeric' });
</script>

{#if unconfigured}
  <EmptyState text={t('waste.add_the_ics_link_of_your')}>
    {#if allowed}
      <button class="btn btn-primary" type="button" onclick={() => (setupOpen = true)}
        >{t('common.set_up')}</button
      >
    {/if}
  </EmptyState>
  {#if setupOpen}
    <WasteSetup
      {waste}
      onclose={() => (setupOpen = false)}
      ondone={() => {
        setupOpen = false;
        reload++;
      }}
    />
  {/if}
{:else if error && !overview}
  <p class="error">{error}</p>
{:else if overview}
  <ul class="types">
    {#each overview.types as t (t.label)}
      <li class:soon={isDueSoon(t, now)} style="--type-color: {t.color}">
        <span class="icon" aria-hidden="true">{t.icon}</span>
        <span class="label">{t.label}</span>
        <span class="when">
          <strong>{dueLabel(t.daysUntil)}</strong>
          {#if t.dates[0] && t.daysUntil > 1}<span class="muted small">{date(t.dates[0])}</span>{/if}
        </span>
      </li>
    {:else}
      <li class="muted">{t('waste.no_upcoming_collections')}</li>
    {/each}
  </ul>
  {#if overview.error}<p class="error small">{overview.error}</p>{/if}
  {#if overview.endsOn}<p class="warning small">
      {t('waste.calendar_ends', { date: date(overview.endsOn) })}
    </p>{/if}
{:else}
  <p class="muted">{t('waste.loading')}</p>
{/if}

<style>
  .warning {
    color: var(--color-warning);
  }
  .types {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 3px;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 2px 6px;
    border-left: 4px solid var(--type-color);
    border-radius: var(--radius-sm);
  }
  li.soon {
    background: color-mix(in srgb, var(--type-color) 22%, transparent);
    font-weight: 600;
  }
  .label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .when {
    display: flex;
    gap: var(--space-1);
    align-items: baseline;
    white-space: nowrap;
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
