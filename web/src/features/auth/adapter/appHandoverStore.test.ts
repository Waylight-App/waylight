import { beforeEach, describe, expect, it } from 'vitest';
import { clearAppHandover, pendingAppHandover, saveAppHandover } from './appHandoverStore';

const handover = { challenge: 'c'.repeat(43), state: 's' };

describe('appHandoverStore', () => {
  beforeEach(() => sessionStorage.clear());

  it('keeps the handover for ten minutes', () => {
    saveAppHandover(handover, 1_000);
    expect(pendingAppHandover(1_000 + 9 * 60_000)).toEqual(handover);
    expect(pendingAppHandover(1_000 + 10 * 60_000)).toBeNull();
  });

  it('forgets an answered handover', () => {
    saveAppHandover(handover);
    clearAppHandover();
    expect(pendingAppHandover()).toBeNull();
  });
});
