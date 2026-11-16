import { describe, expect, it } from 'vitest';
import { filterEntries, type PhoneEntry } from './phonelist';

const entry = (name: string, role = '', team = ''): PhoneEntry => ({
  id: name,
  name,
  role,
  team,
  phone: '1',
  dial: '1',
  email: '',
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
