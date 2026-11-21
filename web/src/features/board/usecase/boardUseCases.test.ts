import { describe, expect, it, vi } from 'vitest';
import { BoardUseCases, type BoardGateway } from './boardUseCases';

describe('BoardUseCases', () => {
  it('trims notes and pins without changing the rest', async () => {
    const g: BoardGateway = {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
      shareGroups: vi.fn(),
    };
    const uc = new BoardUseCases(g);
    await uc.save(null, { text: ' Hi ', pinned: false, expiresOn: '' }, 'private');
    expect(g.create).toHaveBeenCalledWith({ text: 'Hi', pinned: false, expiresOn: '' }, 'private');
    await uc.setPinned(
      { id: '1', text: 'Wi-Fi', pinned: false, expiresOn: '2026-12-01', visibility: 'everyone', mine: true },
      true,
    );
    expect(g.update).toHaveBeenCalledWith('1', { text: 'Wi-Fi', pinned: true, expiresOn: '2026-12-01' });
  });
});
