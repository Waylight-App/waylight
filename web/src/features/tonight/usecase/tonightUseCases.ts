import type { TonightWarning, TonightWeather } from '../domain/tonight';

/** Output port: where the card gets its weather; the composition root adapts the weather feature. */
export interface TonightSource {
  /** The current weather and the next days; null when the widget has no location yet. */
  weather(config: Record<string, unknown>): Promise<TonightWeather | null>;
  /** Active warnings at the widget's location. */
  warnings(config: Record<string, unknown>): Promise<TonightWarning[]>;
  /** Calls back with every newer observation, so the card agrees with the other weather displays; returns the stop function. */
  watch?(config: Record<string, unknown>, onweather: (w: TonightWeather) => void): () => void;
}

/** What one refresh of the card loads. */
export interface TonightSnapshot {
  /** Null without a location. */
  weather: TonightWeather | null;
  warnings: TonightWarning[];
}

/** Application logic of the Tonight card. */
export class TonightUseCases {
  constructor(private readonly source: TonightSource) {}

  /**
   * Loads the weather and the warnings. Warnings are optional, so a failure
   * there leaves the card usable; a failing forecast is the caller's error.
   */
  async snapshot(config: Record<string, unknown>): Promise<TonightSnapshot> {
    const [weather, warnings] = await Promise.all([
      this.source.weather(config),
      this.source.warnings(config).catch((): TonightWarning[] => []),
    ]);
    return { weather, warnings: weather ? warnings : [] };
  }

  /** Follows newer observations of the card's place (#853). */
  watch(config: Record<string, unknown>, onweather: (w: TonightWeather) => void): () => void {
    return this.source.watch?.(config, onweather) ?? (() => {});
  }
}
