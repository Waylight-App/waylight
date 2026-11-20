import { describe, expect, it } from 'vitest';
import { parseLiveNotification } from './live';

describe('live notifications (#583)', () => {
  it('reads the data of a stream event', () => {
    expect(
      parseLiveNotification(
        '{"title":"Disk full","body":"nas","severity":"critical","type":"service","at":"2026-11-20T12:00:00Z"}',
      ),
    ).toEqual({
      title: 'Disk full',
      body: 'nas',
      severity: 'critical',
      type: 'service',
      at: '2026-11-20T12:00:00Z',
    });
  });

  it('falls back to info and refuses broken data', () => {
    expect(parseLiveNotification('{"title":"x","body":"","severity":"loud"}')?.severity).toBe('info');
    expect(parseLiveNotification('not json')).toBeNull();
    expect(parseLiveNotification('{"body":"no title"}')).toBeNull();
    expect(parseLiveNotification('null')).toBeNull();
  });
});
