import { describe, expect, it } from 'vitest';
import { matchKeyword, type OwnKeyword } from './keywords';

const list: OwnKeyword[] = [
  { keyword: 'nas', kind: 'service', target: 's1', label: 'Synology' },
  { keyword: 'müll', kind: 'query', target: 'waste next week', label: '' },
];

describe('matchKeyword', () => {
  it('matches the whole query, ignoring case, accents and spaces around it', () => {
    expect(matchKeyword(list, ' NAS ')?.target).toBe('s1');
    expect(matchKeyword(list, 'mull')?.kind).toBe('query');
  });
  it('leaves longer queries and prefixes to the normal search', () => {
    expect(matchKeyword(list, 'nas status')).toBeNull();
    expect(matchKeyword(list, 'na')).toBeNull();
    expect(matchKeyword(list, '')).toBeNull();
  });
});
