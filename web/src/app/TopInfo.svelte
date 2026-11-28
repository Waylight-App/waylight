<script lang="ts">
  import type { TodosUseCases } from '../features/todos/usecase/todosUseCases';
  import { describe, type Forecast, isDayAt, type WeatherConfig } from '../features/weather/domain/weather';
  import WeatherIcon from '../features/weather/ui/WeatherIcon.svelte';
  import type { WeatherUseCases } from '../features/weather/usecase/weatherUseCases';
  import { intlLocale, t } from '../shared/i18n';
  import { usePolling } from '../shared/refresh';

  /** Props of the centre of the top bar (#189): date, weather and due to-dos. */
  interface Props {
    /** Null when the module is disabled. */
    weather: WeatherUseCases | null;
    /** Location of the first weather widget, or null. */
    location: WeatherConfig | null;
    todos: TodosUseCases | null;
  }
  let { weather, location, todos }: Props = $props();
  const poll = usePolling();

  let now = $state(new Date());
  let forecast = $state<Forecast | null>(null);
  let due = $state<number | null>(null);

  $effect(() => {
    const id = setInterval(() => (now = new Date()), 60_000);
    return () => clearInterval(id);
  });

  // The weather use case shares its forecast with the weather widget (#167)
  // and tells the line about every newer observation (#853), so the line,
  // the tile and the panel always agree.
  $effect(() => {
    const loc = location;
    forecast = null;
    if (!weather || !loc) return;
    let cancelled = false;
    const stop = weather.watch(loc, (f) => (forecast = f));
    weather
      .forecast(loc)
      .then((f) => {
        if (!cancelled) forecast = f;
      })
      .catch(() => {});
    return () => {
      cancelled = true;
      stop();
    };
  });
  const temperature = $derived(
    forecast
      ? {
          value: Math.round(forecast.temperature),
          icon: describe(forecast.code, isDayAt(forecast, now)).icon,
        }
      : null,
  );

  $effect(() => {
    if (!todos) return;
    const uc = todos;
    return poll(async () => {
      const today = new Date();
      const iso = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
      const list = await uc.list(false);
      due = list.filter((x) => x.due !== '' && x.due <= iso).length;
    }, 5 * 60_000);
  });

  const date = $derived(
    now.toLocaleDateString(intlLocale(), { weekday: 'long', day: 'numeric', month: 'long' }),
  );
</script>

<p class="info">
  <span>{date}</span>
  {#if temperature}
    <span class="sep" aria-hidden="true">·</span>
    <span class="temperature"
      ><WeatherIcon icon={temperature.icon} />
      {t('app.greeting_temperature', { t: temperature.value })}</span
    >
  {/if}
  {#if due}
    <span class="sep" aria-hidden="true">·</span>
    <span>{t(due === 1 ? 'app.todo_due_one' : 'app.todos_due', { n: due })}</span>
  {/if}
</p>

<style>
  .info {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: center;
    gap: 0 var(--space-3);
    color: var(--color-text-subtle);
    font-size: var(--fs-xs);
    line-height: var(--lh-xs);
    white-space: nowrap;
  }
  /* A small round dot instead of a typographic one between the items. */
  .sep {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: var(--color-border-strong);
    font-size: 0;
  }
</style>
