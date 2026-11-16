import { describe, expect, it } from 'vitest';
import { filterEntries, groupByTeam, parsePhonelistConfig, type PhoneEntry, toCsv } from './phonelist';

const entry = (name: string, role = '', team = ''): PhoneEntry => ({
  id: name,
  name,
  role,
  team,
  phone: '1',
  dial: '1',
  email: '',
  readOnly: false,
});

describe('filterEntries', () => {
  const list = [entry('Anna Weber', 'Reception', 'Front desk'), entry('Ben Ott', 'Accounting')];

  it('matches every word in name, role or team, ignoring case', () => {
    expect(filterEntries(list, 'front ANNA').map((e) => e.name)).toEqual(['Anna Weber']);
    expect(filterEntries(list, 'accounting').map((e) => e.name)).toEqual(['Ben Ott']);
    expect(filterEntries(list, 'anna accounting')).toEqual([]);
  });

  it('keeps everything for an empty query', () => {
    expect(filterEntries(list, '  ')).toHaveLength(2);
  });
});

describe('groupByTeam', () => {
  it('sorts teams by name and puts people without a team last', () => {
    const groups = groupByTeam([
      entry('A', '', 'Sales'),
      entry('B'),
      entry('C', '', 'Admin'),
      entry('D', '', 'Sales'),
    ]);
    expect(groups.map((g) => g.team)).toEqual(['Admin', 'Sales', '']);
    expect(groups[1]?.entries.map((e) => e.name)).toEqual(['A', 'D']);
  });

  it('reads the widget option', () => {
    expect(parsePhonelistConfig({})).toEqual({ groupByTeam: false });
    expect(parsePhonelistConfig({ groupByTeam: true })).toEqual({ groupByTeam: true });
  });
});

describe('toCsv', () => {
  it('quotes commas, quotes and line breaks and starts with a BOM', () => {
    const e = { ...entry('Weber, Anna', 'Sagt "Hallo"', 'Büro\nNord'), phone: '+49 30 1', email: 'a@b.de' };
    const csv = toCsv([e], ['Name', 'Role', 'Team', 'Phone', 'E-mail']);
    expect(csv.startsWith('\uFEFF')).toBe(true);
    expect(csv.slice(1)).toBe(
      'Name,Role,Team,Phone,E-mail\r\n"Weber, Anna","Sagt ""Hallo""","Büro\nNord",+49 30 1,a@b.de\r\n',
    );
  });
});
