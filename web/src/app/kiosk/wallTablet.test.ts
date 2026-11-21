import { describe, expect, it } from 'vitest';
import { defaultChoice, kioskInput, ready } from './wallTablet';

describe('wall tablet assistant (#680)', () => {
  it('starts with a family screen: template, large text, night mode', () => {
    const c = defaultChoice();
    expect(ready(c)).toBe(true);
    expect(kioskInput(c, 'tab-1')).toEqual({
      name: 'wall-tablet',
      tabs: ['tab-1'],
      disabled: false,
      rotateSeconds: 0,
      nightFrom: '22:00',
      nightTo: '06:30',
      largeText: true,
    });
  });

  it('rotates only through several existing tabs', () => {
    const c = { ...defaultChoice(), source: 'tabs' as const, tabs: ['a'] };
    expect(kioskInput(c).rotateSeconds).toBe(0);
    expect(kioskInput({ ...c, tabs: ['a', 'b'] })).toMatchObject({ tabs: ['a', 'b'], rotateSeconds: 60 });
  });

  it('needs a name and, for existing tabs, at least one', () => {
    expect(ready({ ...defaultChoice(), name: ' ' })).toBe(false);
    expect(ready({ ...defaultChoice(), source: 'tabs' })).toBe(false);
  });

  it('leaves the night window empty when night mode is off', () => {
    expect(kioskInput({ ...defaultChoice(), night: false }, 't')).toMatchObject({
      nightFrom: '',
      nightTo: '',
    });
  });
});
