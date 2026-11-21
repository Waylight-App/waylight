import { describe, expect, it } from 'vitest';
import {
  accessibilityAttributes,
  categoryShown,
  cssVariables,
  defaultAppearance,
  parseTextSize,
} from './appearance';

describe('cssVariables', () => {
  it('is empty for defaults', () => {
    expect(cssVariables(defaultAppearance)).toEqual({});
  });
  it('maps accent and background', () => {
    const vars = cssVariables({
      ...defaultAppearance,
      accentColor: '#112233',
      backgroundImage: 'https://x.y/bg.jpg',
    });
    expect(vars['--color-accent']).toBe('#112233');
    expect(vars['--background-image']).toBe('url("https://x.y/bg.jpg")');
  });
  it('ignores values that could break out of url()', () => {
    expect(
      cssVariables({ ...defaultAppearance, backgroundImage: 'https://x.y/a.png") ; color: red' }),
    ).toEqual({});
  });
});

describe('start-page profile (#281)', () => {
  it('hides home-server and network widgets only for "everyday"', () => {
    expect(categoryShown('everyday', 'everyday')).toBe(true);
    expect(categoryShown('everyday', 'information')).toBe(true);
    expect(categoryShown('everyday', 'server')).toBe(false);
    expect(categoryShown('everyday', 'network')).toBe(false);
    expect(categoryShown('homelab', 'server')).toBe(true);
  });
});

describe('accessibility options (#681)', () => {
  it('sets no root attributes by default', () => {
    expect(accessibilityAttributes(defaultAppearance)).toEqual({
      'data-text-size': undefined,
      'data-contrast': undefined,
      'data-simple': undefined,
    });
  });
  it('maps the options to root attributes', () => {
    expect(
      accessibilityAttributes({
        ...defaultAppearance,
        textSize: 'large',
        highContrast: true,
        simpleView: true,
      }),
    ).toEqual({ 'data-text-size': 'large', 'data-contrast': 'high', 'data-simple': 'true' });
  });
  it('reads an unknown text size as normal', () => {
    expect(parseTextSize('huge')).toBe('normal');
    expect(parseTextSize(undefined)).toBe('normal');
  });
  it('drops the own accent colour and background under high contrast', () => {
    const a = { ...defaultAppearance, accentColor: '#112233', backgroundImage: 'https://x.y/bg.jpg' };
    expect(cssVariables({ ...a, highContrast: true })).toEqual({});
  });
});
