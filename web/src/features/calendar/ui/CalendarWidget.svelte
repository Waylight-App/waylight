<script lang="ts">
  import SetupInPlace from '../../../shared/setup/SetupInPlace.svelte';
  import type { FoundService } from '../../../shared/setup/setup';
  import { calendarSetup } from './calendarSetup';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { ApiError, errorMessage } from '../../../shared/api-client/http';
  import { type AgendaDay, groupByDay, parseCalendarConfig } from '../domain/calendar';
  import type { CalendarUseCases } from '../usecase/calendarUseCases';
  import MonthView from './MonthView.svelte';
  import QuickAddEvent from './QuickAddEvent.svelte';
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import SplitView from './SplitView.svelte';

  /** Props of the calendar widget. */
  interface Props {
    calendar: CalendarUseCases;
    config: Record<string, unknown>;
    /** Services found in the home network for the setup (#590). */
    discover?: () => Promise<FoundService[]>;
  }
  let { calendar, config, discover }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 5 * 60 * 1000;
  const cfg = $derived(parseCalendarConfig(config));
  let days = $state<AgendaDay[]>([]);
  let loaded = $state(false);
  let warning = $state('');
  let error = $state('');
  let unconfigured = $state(false);
  let adding = $state<Date | null>(null);
  let reload = $state(0);
  // svelte-ignore state_referenced_locally
  const setup = calendarSetup(calendar, discover);

  // The agenda list; the month and split views load their own months.
  $effect(() => {
    void reload;
    const c = cfg;
    if (c.view !== 'agenda') return;
    const load = () =>
      calendar
        .events(c.days)
        .then((a) => {
          days = groupByDay(a.events, new Date(), c.days);
          warning = a.error;
          error = '';
          loaded = true;
        })
        .catch((err) => {
          unconfigured = err instanceof ApiError && err.status === 404;
          error = errorMessage(err);
          throw err;
        });
    return poll(load, POLL_MS);
  });

  const dayLabel = (d: Date) => {
    const today = new Date();
    const diff = Math.round(
      (d.getTime() - new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime()) / 86_400_000,
    );
    if (diff === 0) return t('time.today');
    if (diff === 1) return t('time.tomorrow');
    return d.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'short' });
  };
  const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
</script>

{#if unconfigured}
  <!-- Calendars are personal (ADR 0018): every account may connect its own. -->
  <SetupInPlace
    text={t('calendar.connect_a_caldav_calendar_under_settings')}
    descriptor={setup}
    instance={false}
    ondone={() => ((unconfigured = false), (error = ''), reload++)}
  />
{:else if error && !loaded}
  <p class="error">{error}</p>
{:else if cfg.view === 'split'}
  <SplitView
    {calendar}
    weekStart={cfg.weekStart}
    {reload}
    onadd={(d) => (adding = d)}
    onunconfigured={() => (unconfigured = true)}
  />
{:else}
  <button
    class="btn btn-icon add"
    type="button"
    title={t('calendar.new_event')}
    aria-label={t('calendar.new_event')}
    onclick={() => (adding = new Date())}>+</button
  >
  {#if cfg.view === 'month'}
    <MonthView {calendar} weekStart={cfg.weekStart} {reload} onunconfigured={() => (unconfigured = true)} />
  {:else}
    {#if warning}<p class="error small">{warning}</p>{/if}
    {#each days as day (day.key)}
      <section>
        <h4>{dayLabel(day.date)}</h4>
        <ul>
          {#each day.events as e (e.uid + e.start.getTime())}
            <ListRow title={[e.location, e.calendar].filter(Boolean).join(' · ')}>
              {#snippet lead()}<span class="time"
                  >{e.allDay ? t('calendar.all_day_short') : time(e.start)}</span
                >{/snippet}
              <span class="summary" class:colored={e.color !== ''} style:--event-color={e.color || null}
                >{e.summary}</span
              >
            </ListRow>
          {/each}
        </ul>
      </section>
    {:else}
      {#if loaded}<p class="muted">{t('calendar.no_events_in_the_next_days', { days: cfg.days })}</p>{/if}
    {/each}
  {/if}
{/if}
{#if adding}
  <QuickAddEvent {calendar} date={adding} onclose={() => (adding = null)} oncreated={() => reload++} />
{/if}

<style>
  .summary.colored {
    border-left: 3px solid var(--event-color);
    padding-left: var(--space-1);
  }
  .add {
    float: right;
    margin: -4px 0 0 var(--space-1);
    padding: 0 var(--space-2);
    line-height: 1.3;
  }
  h4 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
  }
  section:first-child h4 {
    margin-top: 0;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .time {
    flex: none;
    width: 4.5em;
    font-variant-numeric: tabular-nums;
    color: var(--color-accent-text);
    font-size: var(--fs-xs);
  }
  .small {
    font-size: var(--fs-xs);
  }
</style>
