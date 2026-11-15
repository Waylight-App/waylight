import { describe, expect, it, vi } from 'vitest';
import { type WikiGateway, WikiUseCases } from './wikiUseCases';

const gateway = () =>
  ({
    search: vi.fn(async () => ({ pages: [], failed: [] })),
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
    test: vi.fn(),
  }) satisfies WikiGateway;

describe('WikiUseCases', () => {
  it('does not ask the wikis for one letter', async () => {
    const g = gateway();
    await new WikiUseCases(g).search(' a ');
    expect(g.search).not.toHaveBeenCalled();
    await new WikiUseCases(g).search(' vpn ');
    expect(g.search).toHaveBeenCalledWith('vpn');
  });

  it('creates or updates and trims', async () => {
    const g = gateway();
    const uc = new WikiUseCases(g);
    await uc.save(null, { name: ' H ', kind: 'outline', url: ' https://x ', token: ' k ' });
    expect(g.create).toHaveBeenCalledWith({ name: 'H', kind: 'outline', url: 'https://x', token: 'k' });
    await uc.save('id1', { name: 'H', kind: 'outline', url: 'https://x', token: '' });
    expect(g.update).toHaveBeenCalledWith('id1', expect.objectContaining({ token: '' }));
  });
});
