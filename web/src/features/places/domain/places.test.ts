import { describe, expect, it } from 'vitest';
import { usageList, validateName } from './places';

describe('places', () => {
  it('validates names', () => {
    expect(validateName(' Home ')).toBeNull();
    expect(validateName('  ')).toBe('places.name_empty');
    expect(validateName('x'.repeat(81))).toBe('places.name_too_long');
  });

  it('names the widgets that use a place with their tab', () => {
    const uses = [
      { widgetId: 'w1', tabId: 't1', kind: 'weather' },
      { widgetId: 'w2', tabId: 'gone', kind: 'fuel' },
    ];
    const list = usageList(
      uses,
      (k) => (k === 'weather' ? 'Weather' : 'Fuel prices'),
      (id) => (id === 't1' ? 'Home' : undefined),
    );
    expect(list).toEqual(['Weather (Home)', 'Fuel prices']);
  });
});
