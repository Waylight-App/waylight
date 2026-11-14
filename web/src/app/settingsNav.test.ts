import { describe, expect, it } from 'vitest';
import { groupSections, moveFocus, type SettingsGroup } from './settingsNav';

type S = { id: string; group?: SettingsGroup };

describe('settings navigation (#261)', () => {
  it('groups sections in topic order and drops empty groups', () => {
    const sections: S[] = [
      { id: 'tokens', group: 'account' },
      { id: 'proxmox' },
      { id: 'appearance', group: 'general' },
      { id: 'adguard', group: 'integrations' },
      { id: 'search', group: 'general' },
      { id: 'about', group: 'about' },
    ];
    const names: Record<string, string> = { proxmox: 'Proxmox VE', adguard: 'DNS filter' };
    const groups = groupSections(sections, (s) => names[s.id] ?? s.id, 'en');
    expect(groups.map((g) => g.id)).toEqual(['general', 'integrations', 'account', 'about']);
    expect(groups[0]?.sections.map((s) => s.id)).toEqual(['appearance', 'search']); // given order
    expect(groups[1]?.sections.map((s) => s.id)).toEqual(['adguard', 'proxmox']); // by name, default group
  });

  it('moves the focus with arrow keys, Home and End', () => {
    expect(moveFocus('ArrowDown', 0, 3)).toBe(1);
    expect(moveFocus('ArrowDown', 2, 3)).toBe(0);
    expect(moveFocus('ArrowUp', 0, 3)).toBe(2);
    expect(moveFocus('Home', 2, 3)).toBe(0);
    expect(moveFocus('End', 0, 3)).toBe(2);
    expect(moveFocus('Enter', 0, 3)).toBeNull();
    expect(moveFocus('ArrowDown', 0, 0)).toBeNull();
  });
});
