import { describe, expect, it, vi } from 'vitest';
import { SetupWalk } from './walk.svelte';

describe('SetupWalk', () => {
  it('goes through each kind once, in order', () => {
    const walk = new SetupWalk();
    walk.start(['waste', 'immich', 'waste']);
    expect(walk.current).toBe('waste');
    walk.next();
    expect(walk.current).toBe('immich');
    walk.next();
    expect(walk.active).toBe(false);
    expect(walk.current).toBeUndefined();
  });

  it('calls what was registered once the last setup is over', () => {
    const walk = new SetupWalk();
    const done = vi.fn();
    walk.start(['mail']);
    walk.whenFinished(done);
    expect(done).not.toHaveBeenCalled();
    walk.next();
    walk.next();
    expect(done).toHaveBeenCalledTimes(1);
  });
});
