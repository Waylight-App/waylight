import { describe, expect, it } from 'vitest';
import { stageOf } from './stage';

describe('stageOf', () => {
  it('runs the wizard, then the tour', () => {
    expect(stageOf({ completed: false, step: '' })).toBe('wizard');
    expect(stageOf({ completed: false, step: 'profile' })).toBe('wizard');
    expect(stageOf({ completed: false, step: 'done' })).toBe('tour');
    expect(stageOf({ completed: true, step: '' })).toBe('finished');
  });
});
