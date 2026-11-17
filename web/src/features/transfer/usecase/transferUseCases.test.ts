import { describe, expect, it, vi } from 'vitest';
import { TransferUseCases, type TransferGateway } from './transferUseCases';

describe('TransferUseCases', () => {
  it('rejects foreign files before uploading', async () => {
    const gateway = {
      exportUrl: vi.fn(),
      import: vi.fn(),
      configUrl: vi.fn(),
      applyConfig: vi.fn(),
      managedSections: vi.fn(),
      importForeign: vi.fn(),
    } satisfies TransferGateway;
    await expect(new TransferUseCases(gateway).importFile('{"a":1}')).rejects.toThrow(/not a Waylight/);
    expect(gateway.import).not.toHaveBeenCalled();
  });
});
