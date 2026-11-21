import { describe, expect, it, vi } from 'vitest';
import type { CompactAction } from '../domain/compact';
import { type CompactGateway, CompactUseCases } from './compactUseCases';

const text = (s: string) => ({ key: s, text: s, params: {} });

function setup() {
  const gateway: CompactGateway = { state: vi.fn(), send: vi.fn().mockResolvedValue(undefined) };
  return { gateway, uc: new CompactUseCases(gateway) };
}

describe('compact actions (#598)', () => {
  const action: CompactAction = {
    id: 'habits.toggle',
    label: text('Done: Read'),
    method: 'post',
    path: '/api/habits/h1/toggle',
    body: { date: '2026-11-21' },
    choices: [],
  };

  it('sends the request the action describes', async () => {
    const { uc, gateway } = setup();
    await uc.act(action);
    expect(gateway.send).toHaveBeenCalledWith('POST', '/api/habits/h1/toggle', { date: '2026-11-21' });
  });

  it('sends the body of the chosen chip', async () => {
    const { uc, gateway } = setup();
    const home = { id: 'home', label: text('Home'), body: { status: 'home' }, selected: false };
    await uc.act({ ...action, method: 'PUT', path: '/api/presence/me', choices: [home] }, home);
    expect(gateway.send).toHaveBeenCalledWith('PUT', '/api/presence/me', { status: 'home' });
  });

  it('refuses reads and paths outside the API', async () => {
    const { uc, gateway } = setup();
    await expect(uc.act({ ...action, method: 'GET' })).rejects.toThrow();
    await expect(uc.act({ ...action, path: 'https://example.com/x' })).rejects.toThrow();
    expect(gateway.send).not.toHaveBeenCalled();
  });
});
