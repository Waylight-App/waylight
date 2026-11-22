import { describe, expect, it } from 'vitest';
import { laterSetup, startIndex, templateCards, templateKinds, wizardSteps } from './wizard';

describe('wizardSteps', () => {
  it('asks everything of an owner on a fresh dashboard', () => {
    expect(wizardSteps({ personal: true, owner: true, widgets: true, extra: ['instance'] })).toEqual([
      'welcome',
      'profile',
      'extra:instance',
      'widgets',
      'done',
    ]);
  });

  it('leaves out plugin steps for members and the widgets of a set up dashboard', () => {
    expect(wizardSteps({ personal: true, owner: false, widgets: false, extra: ['instance'] })).toEqual([
      'welcome',
      'profile',
      'done',
    ]);
  });
});

describe('startIndex', () => {
  const steps = ['welcome', 'profile', 'widgets', 'done'];
  it('resumes at the stored step', () => expect(startIndex(steps, 'widgets')).toBe(2));
  it('starts at the beginning without a known step', () => {
    expect(startIndex(steps, '')).toBe(0);
    expect(startIndex(steps, 'gone')).toBe(0);
  });
});

describe('templates', () => {
  const t = (id: string, kinds: string[]) => ({
    id,
    title: id,
    needs: [] as 'location'[],
    widgets: kinds.map((kind, i) => ({ kind, x: i, y: 0, w: 1, h: 1 })),
  });
  it('shows the cards in a fixed order and hides the others', () => {
    expect(templateCards([t('wall', []), t('work', []), t('family', [])]).map((c) => c.id)).toEqual([
      'family',
      'work',
    ]);
  });
  it('lists each kind once', () => {
    expect(templateKinds(t('x', ['todos', 'todos', 'weather']))).toEqual(['todos', 'weather']);
  });
});

describe('laterSetup', () => {
  it('flags widgets that need the user', () => {
    expect(laterSetup('waste', true)).toBe('set_up');
    expect(laterSetup('immich', true)).toBe('connect');
    expect(laterSetup('todos', true)).toBeUndefined();
  });

  it('asks for a place before the weather works', () => {
    expect(laterSetup('weather', false)).toBe('set_up');
    expect(laterSetup('weather', true)).toBeUndefined();
  });
});
