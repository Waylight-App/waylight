<script lang="ts">
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import ListRow from '../../../shared/ui/ListRow.svelte';
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type Departure, minutesUntil, parseTransitConfig } from '../domain/transit';
  import type { TransitUseCases } from '../usecase/transitUseCases';

  /** Props of the transit widget. */
  interface Props {
    transit: TransitUseCases;
    config: Record<string, unknown>;
  }
  let { transit, config }: Props = $props();
  const poll = usePolling();

  const POLL_MS = 60_000;
  const cfg = $derived(parseTransitConfig(config));
  let departures = $state<Departure[]>([]);
  let loaded = $state(false);
  let error = $state('');
  let now = $state(new Date());

  $effect(() => {
    const c = cfg;
    if (!c) return;
    // The shared scheduler pauses while the page is hidden (#117), which
    // matters here: the data comes from a community service.
    const load = () =>
      transit
        .departures(c)
        .then((d) => ((departures = d), (loaded = true), (now = new Date()), (error = '')))
        .catch((err) => {
          error = errorMessage(err);
          throw err;
        });
    const stop = poll(load, POLL_MS);
    // Only re-renders "in x min"; no request.
    const tick = setInterval(() => (now = new Date()), 15_000);
    return () => {
      stop();
      clearInterval(tick);
    };
  });

  const time = (d: Date) => d.toLocaleTimeString(intlLocale(), { hour: '2-digit', minute: '2-digit' });
</script>

{#if !cfg}
  <p class="muted">{t('transit.choose_a_stop_in_the_widget')}</p>
{:else}
  {#if cfg.stopName}<p class="stop muted">{cfg.stopName}</p>{/if}
  {#if error}<p class="error small">{error}</p>{/if}
  <ul class="deps">
    {#each departures as d, i (i)}
      <ListRow nowrap>
        {#snippet lead()}<span class="line" class:cancelled={d.cancelled} title={d.mode}
            ><Icon name="bus" size={16} /> {d.line}</span
          >{/snippet}
        <span class:cancelled={d.cancelled}>{d.headsign}{d.track ? ` · ${d.track}` : ''}</span>
        {#snippet value()}
          <span class="when" title={t('transit.planned', { value: time(d.planned) })}>
            {#if d.cancelled}
              <span class="late">{t('transit.cancelled')}</span>
            {:else}
              {minutesUntil(d, now)}
              {#if d.delayMinutes > 0}<span class="late">+{d.delayMinutes}</span>{/if}
            {/if}
          </span>
        {/snippet}
      </ListRow>
    {:else}
      {#if loaded}<li class="muted">{t('transit.no_departures_in_the_near_future')}</li>{/if}
    {/each}
  </ul>
  <p class="source muted">{t('transit.data_transitous')}</p>
{/if}

<style>
  .stop {
    margin: 0 0 var(--space-1);
    font-size: var(--fs-xs);
  }
  .deps {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .line {
    flex: none;
    min-width: 3.6em;
    font-weight: 600;
  }
  .when {
    white-space: nowrap;
  }
  .late {
    color: var(--color-danger);
    font-size: var(--fs-xs);
  }
  .cancelled {
    text-decoration: line-through;
    color: var(--color-text-subtle);
  }
  .small {
    font-size: var(--fs-xs);
  }
  .source {
    margin: var(--space-1) 0 0;
    font-size: var(--fs-xs);
  }
</style>
