import { describe, expect, it } from 'vitest';
import {
  areaOf,
  groupByArea,
  isSettingsArea,
  moveFocus,
  pushRecent,
  SETTINGS_AREAS,
  type SettingsArea,
} from './settingsNav';
import { OWN_SECTIONS, type SectionViewer, visibleSections } from './settingsCatalog';
import { registeredSections } from './settingsSections';

type S = { id: string; area?: SettingsArea };

describe('settings areas (#600)', () => {
  it('has eight areas in task order', () => {
    expect(SETTINGS_AREAS.map((a) => a.id)).toEqual([
      'account',
      'appearance',
      'notifications',
      'dashboard',
      'connections',
      'people',
      'data',
      'system',
    ]);
  });

  it('groups sections in area order, sorts connections by name and drops empty areas', () => {
    const sections: S[] = [
      { id: 'tokens', area: 'connections' },
      { id: 'proxmox' },
      { id: 'appearance', area: 'appearance' },
      { id: 'adguard' },
      { id: 'feeds', area: 'dashboard' },
      { id: 'search', area: 'dashboard' },
      { id: 'about', area: 'system' },
    ];
    const names: Record<string, string> = {
      proxmox: 'Proxmox VE',
      adguard: 'DNS filter',
      tokens: 'API tokens',
    };
    const areas = groupByArea(sections, (s) => names[s.id] ?? s.id, 'en');
    expect(areas.map((a) => a.id)).toEqual(['appearance', 'dashboard', 'connections', 'system']);
    expect(areas[1]?.sections.map((s) => s.id)).toEqual(['feeds', 'search']); // given order
    expect(areas[2]?.sections.map((s) => s.id)).toEqual(['tokens', 'adguard', 'proxmox']); // by name, default area
  });

  it('puts every section into exactly one known area', () => {
    const all = [...OWN_SECTIONS, ...registeredSections];
    const ids = all.map((s) => s.id);
    expect(new Set(ids).size).toBe(ids.length);
    for (const s of all) expect(isSettingsArea(areaOf(s)), s.id).toBe(true);
  });

  it('keeps the documented mapping of the built-in sections', () => {
    const area = (id: string) => {
      const s = [...OWN_SECTIONS, ...registeredSections].find((x) => x.id === id);
      if (!s) throw new Error(`no section ${id}`);
      return areaOf(s);
    };
    expect(area('account')).toBe('account');
    expect(area('appearance')).toBe('appearance');
    expect(area('notifications')).toBe('notifications');
    expect(area('search')).toBe('dashboard');
    expect(area('todos')).toBe('dashboard');
    expect(area('services')).toBe('connections');
    expect(area('proxmox')).toBe('connections');
    expect(area('tokens')).toBe('connections');
    expect(area('members')).toBe('people');
    expect(area('kiosks')).toBe('people');
    expect(area('backups')).toBe('data');
    expect(area('transfer')).toBe('data');
    expect(area('about')).toBe('system');
    expect(area('device')).toBe('system');
    expect(area('security-log')).toBe('system');
  });

  it('recognises area names only', () => {
    expect(isSettingsArea('people')).toBe(true);
    expect(isSettingsArea('organisation')).toBe(false);
    expect(isSettingsArea(undefined)).toBe(false);
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

  it('keeps the three sections changed last, newest first', () => {
    expect(pushRecent([], 'a')).toEqual(['a']);
    expect(pushRecent(['a', 'b', 'c'], 'd')).toEqual(['d', 'a', 'b']);
    expect(pushRecent(['a', 'b', 'c'], 'c')).toEqual(['c', 'a', 'b']);
  });
});

describe('areas per role (#600)', () => {
  const viewer = (over: Partial<SectionViewer>): SectionViewer => ({
    modules: new Set([
      'appearance',
      'search',
      'notifications',
      'system',
      'backup',
      'transfer',
      'todos',
      'proxmox',
    ]),
    dataKinds: { todos: 'personal-content', search: 'personal-settings' },
    profile: 'homelab',
    localMode: false,
    guest: false,
    member: false,
    admin: true,
    owner: true,
    ...over,
  });
  const areasFor = (v: SectionViewer) =>
    groupByArea(visibleSections([...OWN_SECTIONS, ...registeredSections], v), (s) => s.id).map((a) => a.id);

  it('shows the owner every area with sections', () => {
    expect(areasFor(viewer({}))).toEqual([
      'account',
      'appearance',
      'notifications',
      'dashboard',
      'connections',
      'people',
      'data',
      'system',
    ]);
  });

  it('hides owner-only data from an admin who is not the owner', () => {
    const areas = groupByArea(
      visibleSections([...OWN_SECTIONS, ...registeredSections], viewer({ owner: false })),
      (s) => s.id,
    );
    expect(areas.find((a) => a.id === 'data')).toBeUndefined();
    expect(areas.find((a) => a.id === 'connections')?.sections.map((s) => s.id)).not.toContain('tokens');
  });

  it('gives members their account and personal modules only', () => {
    expect(areasFor(viewer({ member: true, admin: false, owner: false }))).toEqual(['account', 'dashboard']);
  });

  it('gives guests only their account', () => {
    expect(areasFor(viewer({ guest: true, admin: false, owner: false }))).toEqual(['account']);
  });
});
