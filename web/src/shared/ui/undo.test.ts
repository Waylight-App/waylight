import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { offerUndo, UndoQueue } from './undo.svelte';

describe('UndoQueue (#867)', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it('sends the change when the toast goes away', async () => {
    const q = new UndoQueue(1000);
    const commit = vi.fn(() => Promise.resolve());
    q.offer({ text: 'Milk removed', commit, revert: vi.fn() });
    expect(q.current?.text).toBe('Milk removed');
    expect(commit).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1000);
    expect(commit).toHaveBeenCalledOnce();
    expect(q.current).toBeNull();
  });

  it('undo cancels the change instead of sending it', async () => {
    const q = new UndoQueue(1000);
    const commit = vi.fn(() => Promise.resolve());
    const revert = vi.fn();
    q.offer({ text: 'x', commit, revert });
    expect(q.undo()).toBe(true);
    await vi.advanceTimersByTimeAsync(2000);
    expect(commit).not.toHaveBeenCalled();
    expect(revert).toHaveBeenCalledOnce();
    expect(q.undo()).toBe(false);
  });

  it('a new action sends the waiting one first', () => {
    const q = new UndoQueue(1000);
    const first = vi.fn(() => Promise.resolve());
    q.offer({ text: 'a', commit: first, revert: vi.fn() });
    q.offer({ text: 'b', commit: vi.fn(() => Promise.resolve()), revert: vi.fn() });
    expect(first).toHaveBeenCalledOnce();
    expect(q.current?.text).toBe('b');
  });

  it('a failed send brings the item back', async () => {
    const q = new UndoQueue(1000);
    const revert = vi.fn();
    q.offer({ text: 'x', commit: () => Promise.reject(new Error('offline')), revert });
    await q.flush();
    expect(revert).toHaveBeenCalledOnce();
  });

  it('offerUndo keeps the item pending until it is sent or taken back', async () => {
    const q = new UndoQueue(1000);
    const pending = new Set<string>();
    offerUndo(pending, 'a', 'x', () => Promise.resolve(), q);
    expect(pending.has('a')).toBe(true);
    q.undo();
    expect(pending.has('a')).toBe(false);
    offerUndo(pending, 'b', 'x', () => Promise.resolve(), q);
    await q.flush();
    expect(pending.has('b')).toBe(false);
  });
});
