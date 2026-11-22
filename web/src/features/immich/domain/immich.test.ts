import { describe, expect, it } from 'vitest';
import { parseWidgetConfig, thumbnailUrl } from './immich';

describe('immich domain', () => {
  it('builds proxied thumbnail URLs and shows memories by default', () => {
    expect(thumbnailUrl('a/b')).toBe('/api/immich/thumbnails/a%2Fb');
    expect(thumbnailUrl('a', 'c 1')).toBe('/api/immich/thumbnails/a?connection=c%201');
    expect(parseWidgetConfig({})).toEqual({ showMemories: true, connectionId: '' });
    expect(parseWidgetConfig({ showMemories: false, connectionId: 'c1' })).toEqual({
      showMemories: false,
      connectionId: 'c1',
    });
  });
});
