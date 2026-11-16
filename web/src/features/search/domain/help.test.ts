import { describe, expect, it } from 'vitest';
import { filterHelp, groupHelp, type HelpEntry, hintExamples, parseHelpQuery } from './help';

const entries: HelpEntry[] = [
  { id: 'cpu', group: 'home', text: 'How hard the processor works', examples: ['cpu'] },
  { id: 'disk', group: 'home', text: 'How full the disks are', examples: ['disk space'] },
  { id: 'weather', group: 'day', text: 'The weather today', examples: ['weather tomorrow', 'rain'] },
  { id: 'calc', group: 'tools', text: 'Calculate and convert units', examples: ['5 km in mi'] },
];

describe('help query (#479)', () => {
  it('opens on ?, hilfe and help, with an optional filter', () => {
    expect(parseHelpQuery('?')).toBe('');
    expect(parseHelpQuery('  ? ')).toBe('');
    expect(parseHelpQuery('? Wetter')).toBe('wetter');
    expect(parseHelpQuery('hilfe')).toBe('');
    expect(parseHelpQuery('Hilfe Müll')).toBe('mull');
    expect(parseHelpQuery('help disk')).toBe('disk');
  });

  it('leaves other queries alone', () => {
    for (const q of ['helpdesk', 'hilfestellung', '?foo=bar', 'was ist das?', 'cpu', ''])
      expect(parseHelpQuery(q)).toBeNull();
  });
});

describe('help list (#479)', () => {
  it('filters by text and examples, all words, ignoring case and accents', () => {
    expect(filterHelp(entries, '').map((e) => e.id)).toEqual(['cpu', 'disk', 'weather', 'calc']);
    expect(filterHelp(entries, 'weather').map((e) => e.id)).toEqual(['weather']);
    expect(filterHelp(entries, 'RAIN').map((e) => e.id)).toEqual(['weather']);
    expect(filterHelp(entries, 'how full').map((e) => e.id)).toEqual(['disk']);
    expect(filterHelp(entries, 'nothing like this')).toEqual([]);
  });

  it('groups in a fixed order and leaves out empty groups', () => {
    expect(groupHelp(entries).map((s) => [s.group, s.entries.length])).toEqual([
      ['day', 1],
      ['home', 2],
      ['tools', 1],
    ]);
  });

  it('takes hint examples in turns from the groups', () => {
    expect(hintExamples(entries)).toEqual(['weather tomorrow', 'cpu', '5 km in mi', 'disk space']);
    expect(hintExamples([])).toEqual([]);
  });
});
