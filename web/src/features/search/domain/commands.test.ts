import { describe, expect, it } from 'vitest';
import {
  type CommandSpec,
  matchCommands,
  parseCommandQuery,
  permitted,
  scoreCommand,
  type Viewer,
} from './commands';

const member: Viewer = { guest: false, kiosk: false, admin: false, owner: false };
const admin: Viewer = { ...member, admin: true };
const guest: Viewer = { ...member, guest: true };
const kiosk: Viewer = { ...member, kiosk: true };

const dark: CommandSpec = {
  id: 'theme.dark',
  terms: { de: ['dunkel', 'dunkelmodus'], en: ['dark mode'] },
  role: 'member',
  handler: 'theme',
  arg: 'dark',
};
const light: CommandSpec = {
  id: 'theme.light',
  terms: { de: ['hell'], en: ['light mode'] },
  role: 'member',
  handler: 'theme',
  arg: 'light',
};
const addWidget: CommandSpec = {
  id: 'widget.add',
  terms: { de: ['widget hinzufügen'], en: ['add widget'] },
  role: 'member',
  handler: 'addWidget',
  rest: true,
};
const settings: CommandSpec = {
  id: 'settings.open',
  terms: { de: ['einstellungen'], en: ['settings'] },
  role: 'everyone',
  handler: 'settings',
  weight: 1.1,
};
const weather: CommandSpec = {
  id: 'settings.weather',
  terms: { de: ['einstellungen', 'Wetter'], en: ['settings', 'weather'] },
  role: 'everyone',
  handler: 'settings',
  arg: 'weather',
  requires: ['einstellungen', 'settings'],
};
const members: CommandSpec = {
  id: 'settings.members',
  terms: { de: ['einstellungen', 'Mitglieder'], en: ['settings', 'members'] },
  role: 'admin',
  handler: 'settings',
  arg: 'members',
  requires: ['einstellungen', 'settings'],
};
const all = [dark, light, addWidget, settings, weather, members];
const ids = (query: string, viewer = member) => matchCommands(all, query, viewer).map((m) => m.spec.id);

describe('command registry (#476)', () => {
  it('finds commands by German and English words, with typos and prefixes', () => {
    expect(ids('dark mode')).toEqual(['theme.dark']);
    expect(ids('dunkel')).toEqual(['theme.dark']);
    expect(ids('hell')).toEqual(['theme.light']);
    expect(ids('dakr mode')).toEqual(['theme.dark']);
    expect(ids('einstellungen wetter')).toEqual(['settings.weather']);
  });

  it('needs the settings word for generated settings commands', () => {
    expect(ids('wetter')).toEqual([]);
    expect(ids('>wetter')).toEqual(['settings.weather']);
  });

  it('ranks the general command above generated ones with the same word', () => {
    expect(ids('settings')[0]).toBe('settings.open');
  });

  it('passes the words after a rest command on', () => {
    const [m] = matchCommands(all, 'add widget weather', member);
    expect(m?.spec.id).toBe('widget.add');
    expect(m?.rest).toBe('weather');
    // Without rest, an unknown word means no match.
    expect(scoreCommand(dark, 'dark weather')).toBeNull();
  });

  it('respects roles, guests and kiosk displays', () => {
    expect(ids('settings members')).toEqual([]);
    expect(ids('settings members', admin)).toEqual(['settings.members']);
    expect(ids('dark mode', guest)).toEqual([]);
    expect(ids('settings', guest)).toContain('settings.open');
    expect(ids('settings', kiosk)).toEqual([]);
    expect(permitted('owner', admin)).toBe(false);
  });

  it('lists commands after ">" and needs three characters without it', () => {
    expect(parseCommandQuery('> theme')).toEqual({ only: true, text: 'theme' });
    expect(parseCommandQuery('dark')).toEqual({ only: false, text: 'dark' });
    expect(ids('da')).toEqual([]);
    expect(ids('>').length).toBe(5); // all a member may run
  });

  it('applies the boost of earlier choices', () => {
    const boosted = matchCommands(all, '>', member, (id, s) => (id === 'theme.light' ? s + 5 : s));
    expect(boosted[0]?.spec.id).toBe('theme.light');
  });
});
