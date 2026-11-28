import { TonightUseCases, type TonightSource } from '../../features/tonight';
import {
  TONIGHT_DAYS,
  type TonightWarning,
  type TonightWeather,
} from '../../features/tonight/domain/tonight';
import {
  describe,
  type Forecast,
  isDayAt,
  parseWeatherConfig,
  validity,
} from '../../features/weather/domain/weather';
import type { WeatherUseCases } from '../../features/weather/usecase/weatherUseCases';

/**
 * Feeds the Tonight card from the weather feature, so it shares the forecast
 * requests of the weather widget instead of fetching on its own. Only this
 * file knows both features.
 */
export class WeatherTonightSource implements TonightSource {
  constructor(private readonly uc: WeatherUseCases) {}

  async weather(config: Record<string, unknown>): Promise<TonightWeather | null> {
    const cfg = parseWeatherConfig(config);
    if (!cfg) return null;
    return toTonight(await this.uc.forecast({ ...cfg, days: TONIGHT_DAYS }), cfg.label);
  }

  watch(config: Record<string, unknown>, onweather: (w: TonightWeather) => void): () => void {
    const cfg = parseWeatherConfig(config);
    if (!cfg) return () => {};
    return this.uc.watch({ ...cfg, days: TONIGHT_DAYS }, (f) => onweather(toTonight(f, cfg.label)));
  }

  async warnings(config: Record<string, unknown>): Promise<TonightWarning[]> {
    const cfg = parseWeatherConfig(config);
    if (!cfg) return [];
    const now = new Date();
    return (await this.uc.warnings(cfg)).map((w) => ({
      id: w.id,
      title: w.event || w.headline,
      validity: validity(w, now),
      severity: w.severity,
    }));
  }
}

function toTonight(f: Forecast, place: string): TonightWeather {
  const now = describe(f.code, isDayAt(f, new Date()));
  return {
    temperature: f.temperature,
    feelsLike: f.apparentTemperature,
    condition: now.label,
    icon: now.icon,
    place,
    days: f.days.map((d) => {
      const c = describe(d.code);
      return {
        date: d.date,
        icon: c.icon,
        label: c.label,
        max: d.max,
        min: d.min,
        rain: d.precipitationProbability,
      };
    }),
  };
}

const cache = new WeakMap<WeatherUseCases, TonightUseCases>();

/** The card's use cases over the given weather use cases; one instance per container. */
export function tonightFor(weather: WeatherUseCases): TonightUseCases {
  let uc = cache.get(weather);
  if (!uc) cache.set(weather, (uc = new TonightUseCases(new WeatherTonightSource(weather))));
  return uc;
}
