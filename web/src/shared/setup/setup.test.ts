import { describe, expect, it } from 'vitest';
import { addressOf, type SetupDescriptor, stageComplete } from './setup';

const d: SetupDescriptor = {
  service: 'Immich',
  fields: [
    { key: 'url', label: 'setup.address', stage: 'address', type: 'url' },
    { key: 'apiKey', label: 'setup.address', stage: 'credential', type: 'password' },
  ],
  steps: [],
  check: () => Promise.resolve({ lines: [] }),
};

describe('setup descriptor', () => {
  it('normalises the address for links', () => {
    expect(addressOf(d, { url: ' http://192.168.178.20:2283// ' })).toBe('http://192.168.178.20:2283');
  });

  it('knows when a step is filled in', () => {
    expect(stageComplete(d, 'address', { url: 'http://x' })).toBe(true);
    expect(stageComplete(d, 'credential', { url: 'http://x', apiKey: ' ' })).toBe(false);
  });
});
