import { describe, expect, it } from 'vitest';
import { eventLabel, eventTypes, filterByType, isWarning } from './securityEvent';

describe('security events', () => {
  it('maps types to message keys with a fallback', () => {
    expect(eventLabel('login_locked')).toBe('auth.event.login_locked');
    expect(eventLabel('something_new')).toBe('auth.event.other');
  });

  it('highlights suspicious events', () => {
    expect(isWarning('new_device')).toBe(true);
    expect(isWarning('login_succeeded')).toBe(false);
  });
});

describe('security log filter', () => {
  const events = [{ type: 'login_succeeded' }, { type: 'password_changed' }, { type: 'login_succeeded' }];

  it('lists each type that occurs once', () => {
    expect(eventTypes(events)).toEqual(['login_succeeded', 'password_changed']);
  });

  it('filters by type and keeps everything without one', () => {
    expect(filterByType(events, 'login_succeeded')).toHaveLength(2);
    expect(filterByType(events, '')).toHaveLength(3);
  });
});
