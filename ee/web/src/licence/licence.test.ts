import { describe, expect, it, vi } from 'vitest';
import { features } from './features';
import { LicenceApi, statusKey, type Licence } from './licence';
import { texts } from './texts';

const base: Licence = {
  status: 'valid',
  source: 'settings',
  seats: 10,
  accounts: 3,
  features: [],
  unknown: [],
};

describe('licence section', () => {
  it('has a status sentence for every status', () => {
    for (const status of ['missing', 'invalid', 'valid', 'expired', 'over-seats'] as const) {
      expect(texts[statusKey({ ...base, status })]).toBeDefined();
    }
  });

  it('lists every feature once', () => {
    const ids = features.map((f) => f.id);
    expect(new Set(ids).size).toBe(ids.length);
    expect(ids).toHaveLength(12);
  });

  it('trims the key before sending it', async () => {
    const http = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn().mockResolvedValue(base),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    await new LicenceApi(http).install('  KEY  \n');
    expect(http.put).toHaveBeenCalledWith('/api/licence', { key: 'KEY' });
  });
});
