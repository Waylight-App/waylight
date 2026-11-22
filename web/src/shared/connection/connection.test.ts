import { describe, expect, it, vi } from 'vitest';
import {
  connectionIdOf,
  connectionParams,
  connectionIdsOf,
  savedConnections,
  withConnection,
  withConnectionChoice,
  withConnections,
} from './connection';

describe('withConnection', () => {
  it('leaves the path alone without a choice', () => {
    expect(withConnection('/api/immich/overview')).toBe('/api/immich/overview');
    expect(withConnection('/api/immich/overview', '')).toBe('/api/immich/overview');
  });

  it('appends the choice and joins an existing query', () => {
    expect(withConnection('/api/immich/overview', 'a b')).toBe('/api/immich/overview?connection=a%20b');
    expect(withConnection('/api/speedtest/results?hours=24', 'c1')).toBe(
      '/api/speedtest/results?hours=24&connection=c1',
    );
  });

  it('repeats the parameter for several connections', () => {
    expect(withConnections('/api/arr/overview', ['x', 'y'])).toBe(
      '/api/arr/overview?connection=x&connection=y',
    );
  });
});

describe('widget configuration', () => {
  it('reads the choice, ignoring other types', () => {
    expect(connectionIdOf({ connectionId: 'c1' })).toBe('c1');
    expect(connectionIdOf({ connectionId: 7 })).toBe('');
    expect(connectionIdsOf({ connectionIds: ['a', '', 3, 'b'] })).toEqual(['a', 'b']);
    expect(connectionIdsOf({})).toEqual([]);
  });

  it('sets and clears the choice without touching other keys', () => {
    expect(withConnectionChoice({ showMemories: false }, 'c1')).toEqual({
      showMemories: false,
      connectionId: 'c1',
    });
    expect(withConnectionChoice({ connectionId: 'c1', a: 1 }, '')).toEqual({ a: 1 });
  });
});

describe('savedConnections', () => {
  it('offers nothing without a loader or a way to save', () => {
    expect(savedConnections(['immich'], undefined, {}, async () => undefined)).toBeUndefined();
    expect(savedConnections(['immich'], async () => [], {}, undefined)).toBeUndefined();
  });

  it('stores the choice next to the other settings', async () => {
    const save = vi.fn().mockResolvedValue(undefined);
    await savedConnections(['immich'], async () => [], { showMemories: false }, save)?.choose('c1');
    expect(save).toHaveBeenCalledWith({ showMemories: false, connectionId: 'c1' });
  });
});

describe('connectionParams', () => {
  it('names the choice for a compact source, and nothing without one', () => {
    expect(connectionParams({ connectionId: 'c1' })).toEqual({ connection: 'c1' });
    expect(connectionParams({})).toEqual({});
  });
});
