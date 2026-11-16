import { describe, expect, it } from 'vitest';
import { ApiError } from '../../shared/api-client/http';
import type { Container } from '../container';
import { search } from './markets';

const rates = { base: 'EUR', date: '2026-11-13', rates: { EUR: 1, USD: 1.08 }, stale: false };
const sourceWith = (load: () => Promise<typeof rates>) =>
  search.source({ markets: { rates: load } } as unknown as Container, {} as never);

describe('currency conversion in the search (#482)', () => {
  it('answers right away once the rates are loaded, with their date', async () => {
    const s = sourceWith(async () => rates);
    await s.entities?.();
    const a = s.direct?.('108 usd in eur');
    expect(a?.kind).toBe('metric');
    expect(a && 'value' in a ? a.value : '').toMatch(/100/);
    expect(a && 'detail' in a ? a.detail : '').toContain('2026-11-13');
    expect(s.direct?.('5 km in mi')).toBeNull();
  });

  it('says when the rates could not be fetched', async () => {
    const s = sourceWith(async () => {
      throw new ApiError('down', 503, 'unavailable');
    });
    await s.entities?.();
    const a = s.direct?.('50 usd in eur');
    expect(a && 'text' in a ? a.text : '').toMatch(/could not be fetched/);
  });

  it('marks stale rates', async () => {
    const s = sourceWith(async () => ({ ...rates, stale: true }));
    await s.entities?.();
    const a = s.direct?.('10 eur in usd');
    expect(a && 'detail' in a ? a.detail : '').toMatch(/newer rates could not be fetched/);
  });

  it('says why there is no answer when the rates are switched off', async () => {
    const s = sourceWith(async () => {
      throw new ApiError('Currency rates are switched off on this server.', 404, 'not_found');
    });
    await s.entities?.();
    const a = s.direct?.('50 usd in eur');
    expect(a?.kind).toBe('text');
    expect(a && 'text' in a ? a.text : '').toMatch(/switched off/);
    expect(s.direct?.('hello')).toBeNull();
  });
});
