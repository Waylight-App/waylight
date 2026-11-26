import { describe, expect, it } from 'vitest';
import { createHeaderMeta } from './widgetHeaderMeta';

describe('widget header meta (#809)', () => {
  it('starts empty and keeps the last meta', () => {
    const holder = createHeaderMeta();
    expect(holder.meta).toBeNull();
    holder.set({ text: '3 due today' });
    expect(holder.meta).toEqual({ text: '3 due today' });
    holder.set({ text: '1 of 6 failed', tone: 'danger' });
    expect(holder.meta).toEqual({ text: '1 of 6 failed', tone: 'danger' });
  });

  it('treats an empty text as no meta', () => {
    const holder = createHeaderMeta();
    holder.set({ text: 'x' });
    holder.set({ text: '' });
    expect(holder.meta).toBeNull();
    holder.set({ text: 'x' });
    holder.set(null);
    expect(holder.meta).toBeNull();
  });
});
