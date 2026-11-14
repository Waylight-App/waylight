import { describe, expect, it } from 'vitest';
import { Announcer } from './announcer';

describe('Announcer (#113)', () => {
  it('announces each message once per five minutes', () => {
    const said: string[] = [];
    let now = 0;
    const a = new Announcer(
      (m) => said.push(m),
      () => now,
    );
    a.announce('Weather: update failed');
    a.announce('Weather: update failed');
    a.announce('Feeds: update failed');
    a.announce('');
    now = 5 * 60_000;
    a.announce('Weather: update failed');
    expect(said).toEqual(['Weather: update failed', 'Feeds: update failed', 'Weather: update failed']);
  });
  it('repeats a message when asked to (answers to key presses)', () => {
    const said: string[] = [];
    const a = new Announcer((m) => said.push(m));
    a.announce('Clock moved: column 2, row 1', true);
    a.announce('Clock moved: column 2, row 1', true);
    a.announce('', true);
    expect(said).toEqual(['Clock moved: column 2, row 1', 'Clock moved: column 2, row 1']);
  });
});
