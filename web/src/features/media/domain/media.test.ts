import { describe, expect, it } from 'vitest';
import { displayTitle, parseWidgetConfig, posterUrl, providerName } from './media';

describe('media domain', () => {
  it('names episodes with their show', () => {
    expect(displayTitle({ title: 'Pilot', series: 'Show' })).toBe('Show – Pilot');
    expect(displayTitle({ title: 'Film', series: '' })).toBe('Film');
  });
  it('builds proxied poster URLs', () => {
    expect(posterUrl('a b')).toBe('/api/media/posters/a%20b');
    expect(posterUrl('a', 'c1')).toBe('/api/media/posters/a?connection=c1');
  });
  it('shows both parts unless switched off', () => {
    expect(parseWidgetConfig({})).toEqual({ showRecent: true, showSessions: true, connectionId: '' });
    expect(parseWidgetConfig({ showRecent: false })).toEqual({
      showRecent: false,
      showSessions: true,
      connectionId: '',
    });
    expect(providerName('plex')).toBe('Plex');
  });
});
