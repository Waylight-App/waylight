import { describe, expect, it } from 'vitest';
import type { CommandSpec } from '../features/search/domain/commands';
import { buildHelp, type HelpInput } from './searchHelp';

const member = { guest: false, kiosk: false, admin: false, owner: false };
const base: HelpInput = {
  modules: new Set(['weather', 'docker', 'todos', 'search']),
  viewer: member,
  prefixes: { todo: ['t'], note: ['n'], bookmark: ['b'] },
};
const ids = (input: Partial<HelpInput>) => buildHelp({ ...base, ...input }).map((e) => e.id);

describe('buildHelp (#479)', () => {
  it('lists the intents of enabled modules only', () => {
    expect(ids({})).toContain('weather.forecast');
    expect(ids({ modules: new Set(['docker']) })).not.toContain('weather.forecast');
    expect(ids({})).not.toContain('system.cpu');
  });

  it('always lists time and calculator, the web only with the search module', () => {
    expect(ids({ modules: new Set() })).toEqual(['time', 'calc']);
    expect(ids({})).toEqual(expect.arrayContaining(['web', 'bang']));
  });

  it('shows quick capture with the configured prefix, and only for modules that are on', () => {
    const todo = buildHelp(base).find((e) => e.id === 'capture.todo');
    expect(todo?.examples[0]).toMatch(/^t /);
    expect(
      buildHelp({ ...base, prefixes: { todo: ['a'] } }).find((e) => e.id === 'capture.todo')?.examples[0],
    ).toMatch(/^a /);
    expect(ids({})).not.toContain('capture.note');
  });

  it('gives guests and kiosk displays nothing that changes data', () => {
    for (const viewer of [
      { ...member, guest: true },
      { ...member, kiosk: true },
    ]) {
      const list = ids({ viewer });
      expect(list).not.toContain('docker.restart');
      expect(list.some((id) => id.startsWith('capture.'))).toBe(false);
    }
  });

  it('lists the commands the account may run, without generated ones', () => {
    const specs: CommandSpec[] = [
      { id: 'theme.dark', terms: { de: ['dunkel'], en: ['dark'] }, role: 'member', handler: 'theme' },
      {
        id: 'session.logout',
        terms: { de: ['abmelden'], en: ['log out'] },
        role: 'owner',
        handler: 'logout',
      },
      {
        id: 'settings.weather',
        terms: { de: ['wetter'], en: ['weather'] },
        role: 'member',
        handler: 'settings',
        requires: ['settings'],
      },
    ];
    const help = buildHelp({ ...base, commands: { specs: () => specs, title: (s) => `Title ${s.id}` } });
    const commands = help.filter((e) => e.group === 'commands');
    expect(commands).toEqual([
      { id: 'command.theme.dark', group: 'commands', text: 'Title theme.dark', examples: ['dark'] },
    ]);
  });
});
