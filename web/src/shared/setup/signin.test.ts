import { describe, expect, it } from 'vitest';
import { signInAddress, type SetupDescriptor } from './setup';
import { followSignIn, MAX_POLLS, type SignInGateway, SignInTimeout, wait } from './signin';

/** A gateway whose sign-in becomes ready after `after` polls, or fails. */
function fakeGateway(after: number, fail?: Error) {
  const calls: string[] = [];
  let polls = 0;
  const gateway: SignInGateway = {
    start: (service, url) => {
      calls.push(`start ${service} ${url}`);
      return Promise.resolve({ id: 's1', loginUrl: 'https://cloud.example/login/v2/flow/x' });
    },
    status: (id) => {
      polls++;
      calls.push(`status ${id}`);
      if (fail) return Promise.reject(fail);
      return Promise.resolve(polls >= after ? 'ready' : 'pending');
    },
    cancel: (id) => {
      calls.push(`cancel ${id}`);
      return Promise.resolve();
    },
  };
  return { gateway, calls };
}

/** No real waiting in tests; an aborted signal still rejects. */
const instant = (_ms: number, signal: AbortSignal) =>
  signal.aborted ? Promise.reject(signal.reason) : Promise.resolve();

describe('followSignIn', () => {
  it('opens the page and resolves with the ID once the server has the credential', async () => {
    const { gateway, calls } = fakeGateway(3);
    const opened: string[] = [];
    const id = await followSignIn(
      gateway,
      'nextcloud',
      'https://cloud.example',
      (u) => opened.push(u),
      new AbortController().signal,
      instant,
    );
    expect(id).toBe('s1');
    expect(opened).toEqual(['https://cloud.example/login/v2/flow/x']);
    expect(calls).toEqual(['start nextcloud https://cloud.example', 'status s1', 'status s1', 'status s1']);
  });

  it('cancels the sign-in on the server when it fails', async () => {
    const { gateway, calls } = fakeGateway(1, new Error('The sign-in was cancelled.'));
    await expect(
      followSignIn(gateway, 'homeassistant', 'http://ha', () => {}, new AbortController().signal, instant),
    ).rejects.toThrow('cancelled');
    expect(calls.at(-1)).toBe('cancel s1');
  });

  it('stops asking and cancels when the assistant closes', async () => {
    const { gateway, calls } = fakeGateway(1000);
    const controller = new AbortController();
    const sleep = (ms: number, signal: AbortSignal) => {
      if (calls.length >= 3) controller.abort(new Error('closed'));
      return instant(ms, signal);
    };
    await expect(
      followSignIn(gateway, 'nextcloud', 'https://c', () => {}, controller.signal, sleep),
    ).rejects.toThrow('closed');
    expect(calls.filter((c) => c.startsWith('status')).length).toBeLessThan(5);
    expect(calls.at(-1)).toBe('cancel s1');
  });

  it('is bounded even if the server never answers ready', async () => {
    const { gateway, calls } = fakeGateway(Infinity);
    await expect(
      followSignIn(gateway, 'nextcloud', 'https://c', () => {}, new AbortController().signal, instant),
    ).rejects.toBeInstanceOf(SignInTimeout);
    expect(calls.filter((c) => c.startsWith('status')).length).toBe(MAX_POLLS);
    expect(calls.at(-1)).toBe('cancel s1');
  });

  it('does not open the page when the sign-in cannot start', async () => {
    const gateway: SignInGateway = {
      start: () => Promise.reject(new Error('needs PUBLIC_URL')),
      status: () => Promise.resolve('pending'),
      cancel: () => Promise.resolve(),
    };
    const opened: string[] = [];
    await expect(
      followSignIn(
        gateway,
        'homeassistant',
        'http://ha',
        (u) => opened.push(u),
        new AbortController().signal,
      ),
    ).rejects.toThrow('PUBLIC_URL');
    expect(opened).toEqual([]);
  });
});

describe('wait', () => {
  it('rejects as soon as the signal aborts', async () => {
    const controller = new AbortController();
    const pending = wait(60_000, controller.signal);
    controller.abort(new Error('stop'));
    await expect(pending).rejects.toThrow('stop');
  });
});

describe('signInAddress', () => {
  const base: SetupDescriptor = {
    service: 'Nextcloud',
    fields: [{ key: 'url', label: 'setup.address', stage: 'address', type: 'url' }],
    steps: [],
    check: () => Promise.resolve({ lines: [] }),
  };
  const check = () => Promise.resolve({ lines: [] });

  it('is null without a sign-in', () => {
    expect(signInAddress(base, { url: 'https://c' })).toBeNull();
  });

  it('uses the address field by default', () => {
    const d = { ...base, signIn: { service: 'nextcloud' as const, check } };
    expect(signInAddress(d, { url: ' https://cloud.example/ ' })).toBe('https://cloud.example');
    expect(signInAddress(d, { url: '' })).toBeNull();
  });

  it('can derive the address, e.g. from a CalDAV list', () => {
    const d = {
      ...base,
      signIn: {
        service: 'nextcloud' as const,
        check,
        address: (v: Record<string, string>) => (v['url']?.includes('/remote.php/') ? 'https://c/nc' : null),
      },
    };
    expect(signInAddress(d, { url: 'https://c/nc/remote.php/dav/calendars/a/list/' })).toBe('https://c/nc');
    expect(signInAddress(d, { url: 'https://radicale.example/a/list/' })).toBeNull();
  });
});
