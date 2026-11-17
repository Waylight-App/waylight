import { describe as group, expect, it } from 'vitest';
import {
  compass,
  describe,
  type ForecastDay,
  forecastHint,
  parseWeatherConfig,
  placeLabel,
  radarKey,
  validity,
} from './weather';

group('compass (#118)', () => {
  it.each([
    [0, 'n'],
    [22.4, 'n'],
    [22.5, 'ne'],
    [45, 'ne'],
    [90, 'e'],
    [135, 'se'],
    [180, 's'],
    [225, 'sw'],
    [270, 'w'],
    [315, 'nw'],
    [337.4, 'nw'],
    [337.5, 'n'],
    [359.9, 'n'],
  ])('%d° is %s', (degrees, point) => expect(compass(degrees)).toBe(point));

  it('wraps angles outside 0-360', () => {
    expect(compass(360)).toBe('n');
    expect(compass(450)).toBe('e');
    expect(compass(-45)).toBe('nw');
    expect(compass(-360)).toBe('n');
  });
});

group('weather domain', () => {
  it('parses a valid config and clamps days', () => {
    expect(parseWeatherConfig({ lat: 53.1, lon: 8.2, label: 'Home', days: 9 })).toEqual({
      lat: 53.1,
      lon: 8.2,
      label: 'Home',
      days: 5,
      warnings: true,
      pollenRegion: 0,
      pollenTypes: [],
      uvCity: '',
    });
  });
  it('returns null without coordinates', () => {
    expect(parseWeatherConfig({})).toBeNull();
    expect(parseWeatherConfig({ lat: '53', lon: 8 })).toBeNull();
  });
  it('describes WMO codes including night', () => {
    expect(describe(63).label).toBe('Rain');
    expect(describe(0, false).icon).toBe('🌙');
    expect(describe(1234).label).toBe('Unknown');
  });
  it('builds place labels without duplicates', () => {
    expect(placeLabel({ name: 'Berlin', region: 'Berlin', country: 'Germany', lat: 0, lon: 0 })).toBe(
      'Berlin, Germany',
    );
  });

  it('describes warning validity', () => {
    const now = new Date(2026, 10, 11, 20);
    expect(validity({ onset: new Date(2026, 10, 12, 9), expires: null }, now)).toMatch(/^from /);
    expect(validity({ onset: new Date(2026, 10, 11, 19), expires: new Date(2026, 10, 12, 6) }, now)).toMatch(
      /^until /,
    );
    expect(radarKey(new Date(600_000))).toBe(2);
  });
  it('enables warnings by default', () => {
    expect(parseWeatherConfig({ lat: 1, lon: 2 })?.warnings).toBe(true);
    expect(parseWeatherConfig({ lat: 1, lon: 2, warnings: false })?.warnings).toBe(false);
  });
});

group('pollen and UV config (#56)', () => {
  it('reads region, types and city, ignoring unknown types', () => {
    const c = parseWeatherConfig({
      lat: 1,
      lon: 2,
      pollenRegion: 32,
      pollenTypes: ['birch', 'cats', 'hazel'],
      uvCity: 'Hannover',
    });
    expect(c).toMatchObject({ pollenRegion: 32, pollenTypes: ['hazel', 'birch'], uvCity: 'Hannover' });
    expect(parseWeatherConfig({ lat: 1, lon: 2 })).toMatchObject({
      pollenRegion: 0,
      pollenTypes: [],
      uvCity: '',
    });
  });
});

group('forecast hint (#509)', () => {
  const day = (i: number, d: Partial<ForecastDay> = {}): ForecastDay => ({
    date: new Date(2026, 10, 14 + i),
    min: 5,
    max: 12,
    code: 3,
    precipitationProbability: 10,
    ...d,
  });

  it('is empty for an ordinary outlook', () => {
    expect(forecastHint([day(0), day(1), day(2)])).toBeNull();
  });

  it('names the earliest day of the most important kind', () => {
    const days = [day(0, { precipitationProbability: 90 }), day(1, { min: -2 }), day(2, { min: -4 })];
    expect(forecastHint(days)).toMatchObject({ kind: 'frost', day: 1 });
  });

  it('puts a thunderstorm before everything else', () => {
    const days = [day(0, { min: -1 }), day(1), day(2, { code: 95 })];
    expect(forecastHint(days)).toMatchObject({ kind: 'storm', day: 2 });
  });

  it('only looks three days ahead', () => {
    expect(forecastHint([day(0), day(1), day(2), day(3, { max: 33 })])).toBeNull();
  });

  it('warns of heat and of likely rain', () => {
    expect(forecastHint([day(0, { max: 31 })])).toMatchObject({ kind: 'heat', day: 0 });
    expect(forecastHint([day(0), day(1, { precipitationProbability: 70 })])).toMatchObject({
      kind: 'rain',
      day: 1,
    });
  });
});
