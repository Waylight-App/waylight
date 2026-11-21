import { describe, expect, it, vi } from 'vitest';
import type { AttentionItem } from '../domain/attention';
import { AttentionUseCases } from '../usecase/attentionUseCases';
import { AttentionState } from './attentionState.svelte';

const entry: AttentionItem = {
  id: 'widget:w1',
  kind: 'setup',
  integration: 'transit',
  reason: { message: 'Choose a stop for the departures.', messageKey: 'srv_1' },
  since: new Date(),
  action: 'choose',
};

describe('AttentionState', () => {
  it('drops an entry once the server no longer lists it', async () => {
    const list = vi.fn().mockResolvedValueOnce([entry]).mockResolvedValueOnce([]);
    const state = new AttentionState();
    state.attach(new AttentionUseCases({ list, connections: vi.fn() }));
    await state.refresh();
    expect(state.count).toBe(1);
    await state.refresh();
    expect(state.count).toBe(0);
    expect(state.loaded).toBe(true);
  });

  it('shares one request between concurrent refreshes', async () => {
    const list = vi.fn().mockResolvedValue([entry]);
    const state = new AttentionState();
    state.attach(new AttentionUseCases({ list, connections: vi.fn() }));
    await Promise.all([state.refresh(), state.refresh()]);
    expect(list).toHaveBeenCalledTimes(1);
  });

  it('keeps the entries and reports the error when the server fails', async () => {
    const list = vi.fn().mockResolvedValueOnce([entry]).mockRejectedValueOnce(new Error('offline'));
    const state = new AttentionState();
    state.attach(new AttentionUseCases({ list, connections: vi.fn() }));
    await state.refresh();
    await state.check();
    expect(state.count).toBe(1);
    expect(state.error).toBe('offline');
    expect(list).toHaveBeenLastCalledWith(true);
  });

  it('does nothing while detached', async () => {
    const state = new AttentionState();
    await state.refresh();
    expect(state.loaded).toBe(false);
  });
});
