import { describe, expect, it, vi } from 'vitest';
import { AttentionUseCases, type AttentionGateway } from './attentionUseCases';

describe('AttentionUseCases', () => {
  it('asks for kept statuses on list and fresh ones on check', async () => {
    const g = {
      list: vi.fn().mockResolvedValue([]),
      connections: vi.fn().mockResolvedValue([]),
    } satisfies AttentionGateway;
    const uc = new AttentionUseCases(g);
    await uc.list();
    await uc.check();
    expect(g.list.mock.calls).toEqual([[false], [true]]);
    await uc.connections();
    await uc.connections(true);
    expect(g.connections.mock.calls).toEqual([[false], [true]]);
  });
});
