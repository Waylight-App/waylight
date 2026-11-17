import type { Forecast, HealthOptions, HealthReport, Place, WeatherWarning } from '../../domain/weather';
import type { WeatherGateway } from '../../usecase/weatherUseCases';

/** Open-Meteo endpoints, the same the server uses (internal/modules/weather/adapter/gateway). */
export const FORECAST_URL = 'https://api.open-meteo.com/v1/forecast';
export const GEOCODING_URL = 'https://geocoding-api.open-meteo.com/v1/search';
const FORECAST_DAYS = 5;
const MAX_PLACES = 6;

/** Raised for weather data that only a server provides (DWD warnings, pollen and UV). */
export class WeatherNeedsServerError extends Error {
  override name = 'WeatherNeedsServerError';
  constructor() {
    super('This needs a waylight server.');
  }
}

/** The subset of Open-Meteo's forecast response the widget shows. */
interface ForecastPayload {
  current?: {
    temperature_2m?: number;
    apparent_temperature?: number;
    relative_humidity_2m?: number;
    wind_speed_10m?: number;
    wind_direction_10m?: number | null;
    weather_code?: number;
    is_day?: number;
  };
  daily?: {
    time?: string[];
    weather_code?: number[];
    temperature_2m_max?: number[];
    temperature_2m_min?: number[];
    precipitation_probability_max?: (number | null)[];
  };
}

interface GeocodingPayload {
  results?: { name: string; admin1?: string; country?: string; latitude: number; longitude: number }[];
}

/** Turns Open-Meteo's answer into the domain forecast, as the server's openmeteo.go does. */
export function toForecast(p: ForecastPayload, now: Date): Forecast {
  const c = p.current ?? {};
  const d = p.daily ?? {};
  const days = (d.time ?? []).flatMap((day, i) => {
    const code = d.weather_code?.[i];
    const max = d.temperature_2m_max?.[i];
    const min = d.temperature_2m_min?.[i];
    if (code === undefined || max === undefined || min === undefined) return [];
    // Dates come as YYYY-MM-DD in the place's time zone; noon keeps them on that day everywhere.
    return [
      {
        date: new Date(`${day}T12:00:00`),
        min,
        max,
        code,
        precipitationProbability: d.precipitation_probability_max?.[i] ?? 0,
      },
    ];
  });
  const dir = c.wind_direction_10m;
  return {
    temperature: c.temperature_2m ?? 0,
    apparentTemperature: c.apparent_temperature ?? 0,
    humidity: c.relative_humidity_2m ?? 0,
    windSpeed: c.wind_speed_10m ?? 0,
    windDirection: typeof dir === 'number' ? ((dir % 360) + 360) % 360 : null,
    code: c.weather_code ?? 0,
    isDay: c.is_day === 1,
    days,
    updatedAt: now,
  };
}

/**
 * Weather gateway of the browser extension's local mode (ADR 0019). It asks
 * Open-Meteo directly, which allows requests from any origin, so the
 * extension needs no host permission. Coordinates are rounded to two
 * decimals (about 1 km), as on the server. DWD warnings, pollen and UV come
 * from German services through the server and are not available here.
 */
export class LocalWeatherGateway implements WeatherGateway {
  constructor(
    private readonly language: () => string,
    private readonly fetchJson: (url: string) => Promise<unknown> = async (url) => {
      const res = await fetch(url, { credentials: 'omit', referrerPolicy: 'no-referrer' });
      if (!res.ok) throw new Error(`Open-Meteo returned ${res.status}`);
      return res.json();
    },
    private readonly now: () => Date = () => new Date(),
  ) {}

  async forecast(lat: number, lon: number): Promise<Forecast> {
    const q = new URLSearchParams({
      latitude: lat.toFixed(2),
      longitude: lon.toFixed(2),
      current:
        'temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,wind_direction_10m,weather_code,is_day',
      daily: 'weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max',
      timezone: 'auto',
      forecast_days: String(FORECAST_DAYS),
    });
    return toForecast((await this.fetchJson(`${FORECAST_URL}?${q}`)) as ForecastPayload, this.now());
  }

  async searchPlaces(query: string): Promise<Place[]> {
    const q = new URLSearchParams({
      name: query,
      count: String(MAX_PLACES),
      language: this.language(),
      format: 'json',
    });
    const p = (await this.fetchJson(`${GEOCODING_URL}?${q}`)) as GeocodingPayload;
    return (p.results ?? []).map((r) => ({
      name: r.name,
      region: r.admin1 ?? '',
      country: r.country ?? '',
      lat: r.latitude,
      lon: r.longitude,
    }));
  }

  warnings(): Promise<WeatherWarning[]> {
    // No warnings rather than an error: the widget shows the forecast either way.
    return Promise.resolve([]);
  }

  // The rain radar is not offered in local mode (see app/local/modules.ts).
  radarUrl(): string {
    return '';
  }

  health(): Promise<HealthReport> {
    return Promise.reject(new WeatherNeedsServerError());
  }

  healthOptions(): Promise<HealthOptions> {
    return Promise.resolve({ regions: [], cities: [] });
  }
}
