import { describe, expect, it, vi } from 'vitest';
import type { Connection } from '../domain/integrations';
import { ConnectionsUseCases, type ConnectionsGateway } from './connectionsUseCases';

const saved: Connection = {
  id: 'c1',
  service: 'immich',
  name: 'Home',
  url: 'http://immich.lan',
  fields: ['apiKey'],
  values: {},
  usedBy: [],
};

function gateway(over: Partial<ConnectionsGateway> = {}): ConnectionsGateway {
  return {
    services: vi.fn().mockResolvedValue([]),
    list: vi.fn().mockResolvedValue([saved]),
    choices: vi.fn().mockResolvedValue([{ id: 'c1', service: 'immich', name: 'Home' }]),
    add: vi.fn().mockResolvedValue(saved),
    update: vi.fn().mockResolvedValue(saved),
    replaceCredential: vi.fn().mockResolvedValue(saved),
    check: vi.fn().mockResolvedValue(undefined),
    remove: vi.fn().mockResolvedValue(undefined),
    ...over,
  };
}

describe('ConnectionsUseCases', () => {
  it('trims the draft and drops empty credential fields when adding', async () => {
    const g = gateway();
    await new ConnectionsUseCases(g).add({
      service: 'immich',
      name: '  Home ',
      url: ' http://immich.lan ',
      credential: { apiKey: ' k ', note: ' ' },
    });
    expect(g.add).toHaveBeenCalledWith({
      service: 'immich',
      name: 'Home',
      url: 'http://immich.lan',
      credential: { apiKey: 'k' },
    });
  });

  it('keeps the stored credential when renaming', async () => {
    const g = gateway();
    await new ConnectionsUseCases(g).edit(saved, ' Office ', 'http://new.lan');
    expect(g.update).toHaveBeenCalledWith('c1', { name: 'Office', url: 'http://new.lan' });
  });

  it('reports a successful check as ok', async () => {
    expect(await new ConnectionsUseCases(gateway()).check('c1')).toEqual({ kind: 'ok' });
  });

  it('turns a refused check into a failing state with the reason', async () => {
    const g = gateway({ check: vi.fn().mockRejectedValue(new Error('refused')) });
    const state = await new ConnectionsUseCases(g).check('c1');
    expect(state.kind).toBe('failing');
  });

  it('lists the choices of one service', async () => {
    const g = gateway();
    expect(await new ConnectionsUseCases(g).choices('immich')).toEqual([
      { id: 'c1', service: 'immich', name: 'Home' },
    ]);
    expect(g.choices).toHaveBeenCalledWith('immich');
  });
});
