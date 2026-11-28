<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { usePolling } from '../../../shared/refresh';
  import WidgetHeader from '../../../shared/ui/WidgetHeader.svelte';
  import { describe, type Forecast, isDayAt, parseWeatherConfig } from '../domain/weather';
  import type { WeatherUseCases } from '../usecase/weatherUseCases';
  import WeatherIcon from './WeatherIcon.svelte';

  /**
   * Props of the compact weather view for small and strip tiles (#853). It
   * reads the same observation as the detail panel and the top bar, so the
   * three never disagree on the temperature or the day/night icon.
   */
  interface Props {
    weather: WeatherUseCases;
    config: Record<string, unknown>;
    /** One line for strip tiles. */
    strip?: boolean;
  }
  let { weather, config, strip = false }: Props = $props();
  const poll = usePolling();

  /** The backend refreshes every 15 minutes; polling more often is pointless. */
  const POLL_MS = 10 * 60 * 1000;
  const cfg = $derived(parseWeatherConfig(config));
  let forecast = $state<Forecast | null>(null);
  let failed = $state(false);
  let clock = $state(new Date());

  $effect(() => {
    const c = cfg;
    if (!c) return;
    const stop = weather.watch(c, (f) => (forecast = f));
    const stopPoll = poll(
      () =>
        weather.forecast(c).then(
          (f) => ((forecast = f), (failed = false)),
          (err) => {
            failed = true;
            throw err;
          },
        ),
      POLL_MS,
    );
    return () => (stop(), stopPoll());
  });

  $effect(() => {
    const id = setInterval(() => (clock = new Date()), 60_000);
    return () => clearInterval(id);
  });

  const now = $derived(forecast ? describe(forecast.code, isDayAt(forecast, clock)) : null);
</script>

<div class="compact" class:strip data-testid="weather-compact">
  <div class="label"><WidgetHeader title={t('widget.weather.title')} /></div>
  {#if forecast && now}
    <span class="state"
      ><span class="icon" data-icon={now.icon}><WeatherIcon icon={now.icon} /></span>{Math.round(
        forecast.temperature,
      )}°</span
    >
    <span class="detail">{now.label}</span>
  {:else if failed}
    <span class="detail">{t('compact.unavailable')}</span>
  {:else}
    <span class="detail">{t('compact.loading')}</span>
  {/if}
</div>

<style>
  .compact {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: var(--space-1);
    min-width: 0;
  }
  .compact > * {
    flex: none;
  }
  .compact.strip {
    flex-direction: row;
    align-items: center;
    gap: var(--space-2);
  }
  .label {
    min-width: 0;
  }
  .strip .label {
    max-width: 50%;
  }
  .state,
  .detail {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .state {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-md);
    line-height: var(--lh-md);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .strip .state {
    font-size: inherit;
    line-height: var(--lh-sm);
  }
  .icon {
    display: inline-flex;
    color: var(--color-accent-text);
  }
  /* Line icons are 16 px everywhere. */
  .icon :global(svg) {
    width: 16px;
    height: 16px;
  }
  .detail {
    color: var(--color-text-subtle);
    font-size: var(--fs-sm);
  }
</style>
