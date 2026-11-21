import { describe, expect, it, vi } from 'vitest';
import { DiscoveryUseCases } from './discoveryUseCases';

describe('DiscoveryUseCases', () => {
  it('passes the services on and turns a refusal into an empty list', async () => {
    const found = [{ url: 'http://172.18.0.5:2283', hint: 'Docker container immich_server' }];
    expect(await new DiscoveryUseCases({ find: vi.fn().mockResolvedValue(found) }).find('immich')).toEqual(
      found,
    );
    expect(
      await new DiscoveryUseCases({ find: vi.fn().mockRejectedValue(new Error('403')) }).find('immich'),
    ).toEqual([]);
  });
});
