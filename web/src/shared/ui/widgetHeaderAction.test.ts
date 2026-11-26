import { describe, expect, it } from 'vitest';
import { createHeaderAction } from './widgetHeaderAction';

describe('widget header action (#809)', () => {
  it('holds the last action until it is cleared', () => {
    const holder = createHeaderAction();
    expect(holder.action).toBeNull();
    const action = { icon: 'plus' as const, label: 'New note', onclick: () => {} };
    holder.set(action);
    expect(holder.action).toBe(action);
    holder.set(null);
    expect(holder.action).toBeNull();
  });
});
