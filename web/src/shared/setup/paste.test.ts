import { describe, expect, it } from 'vitest';
import { followPaste, type PasteGateway, PasteTimeout, pasteLink } from './paste';

function fakeGateway(readyAfter: number, expiresAt = new Date(10_000)) {
  const calls: string[] = [];
  let takes = 0;
  const gateway: PasteGateway = {
    start: () => Promise.resolve({ code: 'c1', expiresAt }),
    take: () => {
      takes++;
      calls.push('take');
      return Promise.resolve(takes >= readyAfter ? 'the-key' : null);
    },
    cancel: (code) => {
      calls.push(`cancel ${code}`);
      return Promise.resolve();
    },
  };
  return { gateway, calls };
}

const instant = (_ms: number, signal: AbortSignal) =>
  signal.aborted ? Promise.reject(signal.reason) : Promise.resolve();

describe('followPaste', () => {
  it('shows the code and resolves with the key once the phone sent it', async () => {
    const { gateway, calls } = fakeGateway(2);
    const shown: string[] = [];
    const value = await followPaste(
      gateway,
      (s) => shown.push(s.code),
      new AbortController().signal,
      instant,
      () => 0,
    );
    expect(value).toBe('the-key');
    expect(shown).toEqual(['c1']);
    expect(calls).toEqual(['take', 'take']);
  });

  it('cancels the code and rejects when it expires', async () => {
    const { gateway, calls } = fakeGateway(1000);
    let t = 0;
    const clock = () => (t += 4000);
    await expect(
      followPaste(gateway, () => {}, new AbortController().signal, instant, clock),
    ).rejects.toBeInstanceOf(PasteTimeout);
    expect(calls.at(-1)).toBe('cancel c1');
  });

  it('cancels the code when the assistant closes', async () => {
    const { gateway, calls } = fakeGateway(1000);
    const controller = new AbortController();
    const sleep = (_ms: number, signal: AbortSignal) => {
      controller.abort(new Error('closed'));
      return signal.aborted ? Promise.reject(signal.reason) : Promise.resolve();
    };
    await expect(
      followPaste(
        gateway,
        () => {},
        controller.signal,
        sleep,
        () => 0,
      ),
    ).rejects.toThrow('closed');
    expect(calls).toEqual(['cancel c1']);
  });
});

describe('pasteLink', () => {
  it('carries the address and the code', () => {
    expect(pasteLink('https://waylight.example', 'abc')).toBe(
      'waylight://paste?server=https%3A%2F%2Fwaylight.example&code=abc',
    );
  });
});
