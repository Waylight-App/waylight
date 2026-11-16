import { describe, expect, it } from 'vitest';
import { calculate } from '../features/search/domain/calc/calc';
import { SmartSearch } from '../features/search/usecase/smartSearch';
import { toolsSource } from './searchTools';

let n = 0;
const env = {
  // Different bytes on every call, so regenerating gives a new value.
  random: (len: number) => Uint8Array.from({ length: len }, () => (n = (n * 97 + 13) % 256)),
  now: () => new Date('2026-11-16T12:00:00Z'),
};

const search = () =>
  new SmartSearch({
    intents: [],
    sources: [toolsSource(env)],
    canAct: false,
    askBack: () => {
      throw new Error('unused');
    },
  });

describe('toolsSource', () => {
  it('loads a tool on demand and then answers directly, so Enter copies what is shown', async () => {
    const s = search();
    expect(s.direct('uuid')).toEqual([]);
    const [answer] = await s.answer('uuid', new AbortController().signal);
    expect(answer).toMatchObject({ kind: 'metric', intent: 'tools.uuid' });
    expect(s.direct('uuid')[0]?.copy).toBe(answer?.copy);
    // A second answer() does not generate a different value behind the user's back.
    expect((await s.answer('uuid', new AbortController().signal))[0]?.copy).toBe(answer?.copy);
  });

  it('makes a new password on request and copies the new one', async () => {
    const s = search();
    const [first] = await s.answer('pw 12', new AbortController().signal);
    const next = await first?.regenerate?.();
    expect(next?.copy).toHaveLength(12);
    expect(next?.copy).not.toBe(first?.copy);
    expect(s.direct('pw 12')[0]?.copy).toBe(next?.copy);
  });

  it('shows a colour with a swatch and a subnet as a list', async () => {
    const s = search();
    const [c] = await s.answer('#1e90ff', new AbortController().signal);
    expect(c).toMatchObject({ swatch: '#1e90ff', copy: '#1e90ff' });
    const [net] = await s.answer('192.168.178.0/24', new AbortController().signal);
    expect(net).toMatchObject({ kind: 'list', copy: '192.168.178.0/24' });
  });

  it('stays silent for other queries and invalid input', async () => {
    const s = search();
    expect(await s.answer('weather tomorrow', new AbortController().signal)).toEqual([]);
    expect(await s.answer('rgb(300, 0, 0)', new AbortController().signal)).toEqual([]);
  });

  it('does not compete with the calculator over a bare timestamp', () => {
    expect(calculate('1700000000')).toBeNull();
  });
});
