import { describe, expect, it } from 'vitest';
import { LocalWeatherGateway, WeatherNeedsServerError, toForecast } from './localWeatherGateway';

const payload = {
  current: {
    temperature_2m: 7.4,
    apparent_temperature: 4.9,
    relative_humidity_2m: 81,
    wind_speed_10m: 14.2,
    wind_direction_10m: -90,
    weather_code: 3,
    is_day: 1,
  },
  daily: {
    time: ['2026-11-16', '2026-11-17'],
    weather_code: [3, 61],
    temperature_2m_max: [9.1, 8],
    temperature_2m_min: [3.2, 4.5],
    precipitation_probability_max: [10, null],
  },
};

describe('toForecast (mirrors openmeteo.go)', () => {
  it('maps current conditions and days', () => {
    const f = toForecast(payload, new Date('2026-11-16T07:00:00Z'));
    expect(f).toMatchObject({
      temperature: 7.4,
      apparentTemperature: 4.9,
      humidity: 81,
      code: 3,
      isDay: true,
    });
    expect(f.windDirection).toBe(270);
    expect(f.days).toHaveLength(2);
    expect(f.days[1]).toMatchObject({ code: 61, min: 4.5, max: 8, precipitationProbability: 0 });
  });

  it('keeps an unknown wind direction apart from north', () => {
    expect(toForecast({ current: { wind_direction_10m: null } }, new Date()).windDirection).toBeNull();
    expect(toForecast({ current: { wind_direction_10m: 0 } }, new Date()).windDirection).toBe(0);
  });
});

describe('LocalWeatherGateway', () => {
  it('asks Open-Meteo with coordinates rounded like the server', async () => {
    const urls: string[] = [];
    const gateway = new LocalWeatherGateway(
      () => 'de',
      (url) => (urls.push(url), Promise.resolve(payload)),
    );
    await gateway.forecast(52.520008, 13.404954);
    const u = new URL(urls[0] ?? '');
    expect(u.origin).toBe('https://api.open-meteo.com');
    expect(u.searchParams.get('latitude')).toBe('52.52');
    expect(u.searchParams.get('longitude')).toBe('13.40');
    expect(u.searchParams.get('forecast_days')).toBe('5');
  });

  it('searches places in the UI language', async () => {
    const urls: string[] = [];
    const gateway = new LocalWeatherGateway(
      () => 'de',
      (url) => (
        urls.push(url),
        Promise.resolve({
          results: [
            { name: 'Berlin', admin1: 'Berlin', country: 'Deutschland', latitude: 52.52, longitude: 13.41 },
          ],
        })
      ),
    );
    expect(await gateway.searchPlaces('Berlin')).toEqual([
      { name: 'Berlin', region: 'Berlin', country: 'Deutschland', lat: 52.52, lon: 13.41 },
    ]);
    expect(new URL(urls[0] ?? '').searchParams.get('language')).toBe('de');
  });

  it('has no warnings, radar, pollen or UV without a server', async () => {
    const gateway = new LocalWeatherGateway(
      () => 'en',
      () => Promise.reject(new Error('no network in this test')),
    );
    expect(await gateway.warnings()).toEqual([]);
    expect(gateway.radarUrl()).toBe('');
    expect(await gateway.healthOptions()).toEqual({ regions: [], cities: [] });
    await expect(gateway.health()).rejects.toBeInstanceOf(WeatherNeedsServerError);
  });
});
