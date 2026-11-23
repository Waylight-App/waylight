import { describe, expect, it } from 'vitest';
import { calculate } from '../../search/domain/calc/calc';
import {
  exampleSetOf,
  laterSetup,
  searchExamples,
  startIndex,
  suggestionsOf,
  suggestionTiles,
  templateCards,
  templateKinds,
  wizardSteps,
} from './wizard';

describe('wizardSteps', () => {
  it('asks everything of an owner on a fresh dashboard', () => {
    expect(wizardSteps({ personal: true, owner: true, widgets: true, extra: ['instance'] })).toEqual([
      'welcome',
      'profile',
      'appearance',
      'extra:instance',
      'widgets',
      'search',
      'done',
    ]);
  });

  it('leaves out plugin steps for members and the widgets of a set up dashboard', () => {
    expect(wizardSteps({ personal: true, owner: false, widgets: false, extra: ['instance'] })).toEqual([
      'welcome',
      'profile',
      'appearance',
      'search',
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

describe('suggestions', () => {
  it('preselects the core of a template and offers more that fit', () => {
    const s = suggestionsOf('homelab');
    expect(s.selected).toEqual(['clock', 'todos', 'bookmarks', 'services']);
    expect(s.offered).not.toContain('clock');
  });
  it('offers the single widgets, none selected, without a template', () => {
    expect(suggestionsOf('empty').selected).toEqual([]);
    expect(suggestionsOf('').offered).toContain('weather');
  });
  it('draws the available suggestions and appends other chosen kinds', () => {
    const s = { selected: ['a', 'b'], offered: ['c', 'd'] };
    expect(suggestionTiles(s, ['a', 'x'], new Set(['a', 'b', 'd', 'x']))).toEqual(['a', 'b', 'd', 'x']);
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

describe('searchExamples', () => {
  it('follows the template and leaves out examples whose module is off', () => {
    expect(exampleSetOf('family')).toBe('family');
    expect(exampleSetOf('empty')).toBe('home');
    const all = new Set(['calendar', 'birthdays', 'waste', 'weather', 'shopping']);
    expect(searchExamples('en', 'family', all)[0]).toBe('What is on today?');
    expect(searchExamples('de', 'family', new Set(['waste']))).toEqual([
      'Wann kommt die Biotonne?',
      '19 % von 240',
      '5 km in mi',
    ]);
  });

  it('keeps the calculator examples answerable', () => {
    for (const lang of ['de', 'en'] as const)
      for (const q of searchExamples(lang, 'home', new Set()).slice(0, 2))
        expect(calculate(q)).not.toBeNull();
  });
});
