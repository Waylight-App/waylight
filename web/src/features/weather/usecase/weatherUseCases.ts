import {
  type Forecast,
  type HealthOptions,
  type HealthReport,
  type Place,
  wantsHealth,
  type WeatherConfig,
  type WeatherWarning,
} from '../domain/weather';

/** Output port: weather backend. */
export interface WeatherGateway {
  forecast(lat: number, lon: number): Promise<Forecast>;
  searchPlaces(query: string): Promise<Place[]>;
  warnings(lat: number, lon: number): Promise<WeatherWarning[]>;
  radarUrl(lat: number, lon: number, key: number): string;
  health(region: number, city: string, types: string[]): Promise<HealthReport>;
  healthOptions(): Promise<HealthOptions>;
}

/** How long a loaded forecast is shared between callers. */
export const FORECAST_SHARE_MS = 60_000;

/** Minimum length of a place search. */
export const MIN_PLACE_QUERY = 2;

/** The forecast with the days a widget shows. */
function trim(f: Forecast, cfg: Pick<WeatherConfig, 'days'>): Forecast {
  return { ...f, days: f.days.slice(0, cfg.days) };
}

/** Application logic of the weather feature. */
export class WeatherUseCases {
  /** Recent forecasts per place, so the widget and the greeting line (#167) share one request. */
  private readonly recent = new Map<string, { at: number; forecast: Promise<Forecast> }>();
  /** The newest observation per place and who shows it (#853). */
  private readonly latest = new Map<string, Forecast>();
  private readonly watchers = new Map<string, Set<(f: Forecast) => void>>();

  constructor(
    private readonly gateway: WeatherGateway,
    private readonly now: () => number = Date.now,
  ) {}

  /** Loads the forecast for a configured widget and trims it to the configured days. */
  async forecast(cfg: WeatherConfig): Promise<Forecast> {
    const key = `${cfg.lat},${cfg.lon}`;
    let hit = this.recent.get(key);
    if (!hit || this.now() - hit.at > FORECAST_SHARE_MS) {
      hit = { at: this.now(), forecast: this.gateway.forecast(cfg.lat, cfg.lon) };
      this.recent.set(key, hit);
      hit.forecast.catch(() => this.recent.delete(key)); // a failure is not shared
    }
    return trim(this.keep(key, await hit.forecast), cfg);
  }

  /**
   * Calls back with every newer observation of the place, whoever loaded it,
   * so the tile, the detail panel and the top bar always show the same one
   * (#853). Returns the function that stops watching.
   */
  watch(cfg: WeatherConfig, onforecast: (f: Forecast) => void): () => void {
    const key = `${cfg.lat},${cfg.lon}`;
    const cb = (f: Forecast) => onforecast(trim(f, cfg));
    let set = this.watchers.get(key);
    if (!set) this.watchers.set(key, (set = new Set()));
    set.add(cb);
    return () => set.delete(cb);
  }

  /** Keeps the newest observation of a place; an older answer that arrives late gives way to it. */
  private keep(key: string, f: Forecast): Forecast {
    const known = this.latest.get(key);
    if (known && known.updatedAt >= f.updatedAt) return known;
    this.latest.set(key, f);
    for (const cb of this.watchers.get(key) ?? []) cb(f);
    return f;
  }

  /** Searches places; short queries return no results without a request. */
  searchPlaces(query: string): Promise<Place[]> {
    const q = query.trim();
    if (q.length < MIN_PLACE_QUERY) return Promise.resolve([]);
    return this.gateway.searchPlaces(q);
  }

  /** Official warnings for a configured widget (empty when disabled). */
  warnings(cfg: WeatherConfig): Promise<WeatherWarning[]> {
    return cfg.warnings ? this.gateway.warnings(cfg.lat, cfg.lon) : Promise.resolve([]);
  }

  /** Pollen and UV index of a widget (#56); null without a region or city, so nothing is requested. */
  health(cfg: WeatherConfig): Promise<HealthReport | null> {
    if (!wantsHealth(cfg)) return Promise.resolve(null);
    return this.gateway.health(cfg.pollenRegion, cfg.uvCity, cfg.pollenTypes);
  }

  /** Selectable pollen regions and UV cities for the widget settings. */
  healthOptions(): Promise<HealthOptions> {
    return this.gateway.healthOptions();
  }

  /** URL of the radar image; key changes every five minutes. */
  radarUrl(cfg: Pick<WeatherConfig, 'lat' | 'lon'>, key: number): string {
    return this.gateway.radarUrl(cfg.lat, cfg.lon, key);
  }
}
