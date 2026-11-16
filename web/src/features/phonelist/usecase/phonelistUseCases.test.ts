import { describe, expect, it, vi } from 'vitest';
import { PhonelistUseCases, type PhonelistGateway } from './phonelistUseCases';

describe('PhonelistUseCases', () => {
  it('trims the input and adds or updates', async () => {
    const gateway: PhonelistGateway = {
      list: vi.fn(),
      add: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
    };
    const uc = new PhonelistUseCases(gateway);
    await uc.save({ name: ' Anna ', role: '', phone: ' 1 ', email: '', team: '' });
    expect(gateway.add).toHaveBeenCalledWith({ name: 'Anna', role: '', phone: '1', email: '', team: '' });
    await uc.save({ name: 'Anna', role: 'x', phone: '1', email: '', team: '' }, 'e1');
    expect(gateway.update).toHaveBeenCalledWith('e1', {
      name: 'Anna',
      role: 'x',
      phone: '1',
      email: '',
      team: '',
    });
  });
});
