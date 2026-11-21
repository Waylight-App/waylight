import { describe, expect, it } from 'vitest';
import { addressOf, initialValues, type SetupDescriptor, stageComplete } from './setup';

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

  it('preselects the first option and skips optional fields', () => {
    const choice: SetupDescriptor = {
      ...d,
      fields: [
        {
          key: 'app',
          label: 'setup.address',
          stage: 'address',
          type: 'choice',
          options: [
            { value: 'sonarr', label: 'Sonarr' },
            { value: 'radarr', label: 'Radarr' },
          ],
        },
        { key: 'password', label: 'setup.address', stage: 'credential', type: 'password', optional: true },
      ],
    };
    expect(initialValues(choice)).toEqual({ app: 'sonarr' });
    expect(stageComplete(choice, 'address', initialValues(choice))).toBe(true);
    expect(stageComplete(choice, 'credential', {})).toBe(true);
  });
});
