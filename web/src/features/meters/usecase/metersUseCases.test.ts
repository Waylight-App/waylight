import { describe, expect, it, vi } from 'vitest';
import type { Meter } from '../domain/meters';
import { MetersUseCases, type MetersGateway } from './metersUseCases';

const meter = { id: 'm1' } as Meter;
const input = { name: ' Power ', kind: 'electricity' as const, unit: ' kWh ' };

const gateway = (): MetersGateway => ({
  list: vi.fn(),
  detail: vi.fn(),
  create: vi.fn().mockResolvedValue(meter),
  update: vi.fn().mockResolvedValue(meter),
  remove: vi.fn(),
  addReading: vi.fn().mockResolvedValue({ day: '2026-11-01', value: 5 }),
  removeReading: vi.fn(),
  shareGroups: vi.fn(),
});

describe('MetersUseCases', () => {
  it('creates a meter with trimmed fields and changes an existing one', async () => {
    const g = gateway();
    const uc = new MetersUseCases(g);
    await uc.save(null, input, 'everyone');
    expect(g.create).toHaveBeenCalledWith({ name: 'Power', kind: 'electricity', unit: 'kWh' }, 'everyone');
    await uc.save('m1', input);
    expect(g.update).toHaveBeenCalledWith('m1', expect.objectContaining({ name: 'Power' }), undefined);
  });

  it('sends a reading with its day and value', async () => {
    const g = gateway();
    await new MetersUseCases(g).addReading('m1', { day: '2026-11-01', value: 5 });
    expect(g.addReading).toHaveBeenCalledWith('m1', { day: '2026-11-01', value: 5 });
  });
});
