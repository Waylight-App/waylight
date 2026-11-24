import { describe, expect, it } from 'vitest';
import type { TonightWeather } from '../domain/tonight';
import { type TonightSource, TonightUseCases } from './tonightUseCases';

const sunny: TonightWeather = { temperature: 9, condition: 'Clear', icon: 'clear-day', place: '', days: [] };
const warning = { id: 'w', title: 'Frost', validity: '', severity: 'minor' } as const;

const source = (over: Partial<TonightSource>): TonightSource => ({
  weather: () => Promise.resolve(sunny),
  warnings: () => Promise.resolve([warning]),
  ...over,
});

describe('TonightUseCases.snapshot', () => {
  it('combines the weather and the warnings', async () => {
    const s = await new TonightUseCases(source({})).snapshot({});
    expect(s).toEqual({ weather: sunny, warnings: [warning] });
  });

  it('keeps the card usable when the warnings fail', async () => {
    const uc = new TonightUseCases(source({ warnings: () => Promise.reject(new Error('down')) }));
    expect((await uc.snapshot({})).warnings).toEqual([]);
  });

  it('shows no warnings without a location', async () => {
    const s = await new TonightUseCases(source({ weather: () => Promise.resolve(null) })).snapshot({});
    expect(s).toEqual({ weather: null, warnings: [] });
  });

  it('passes a failing forecast on', async () => {
    const uc = new TonightUseCases(source({ weather: () => Promise.reject(new Error('offline')) }));
    await expect(uc.snapshot({})).rejects.toThrow('offline');
  });
});
