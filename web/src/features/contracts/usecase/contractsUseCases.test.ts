import { describe, expect, it, vi } from 'vitest';
import { ContractsUseCases, type ContractsGateway } from './contractsUseCases';

describe('ContractsUseCases', () => {
  it('sends the cost in cents', async () => {
    const g: ContractsGateway = {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
      shareGroups: vi.fn(),
    };
    const input = {
      name: ' Gym ',
      provider: '',
      cost: '34,90',
      period: 'monthly' as const,
      termEnd: '2027-03-31',
      noticeDays: 90,
      renewMonths: 12,
      remindDays: 14,
      note: '',
    };
    await new ContractsUseCases(g).save(null, input);
    expect(g.create).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'Gym', costCents: 3490 }),
      undefined,
    );
  });
});
