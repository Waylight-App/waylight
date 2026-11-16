import { describe, expect, it } from 'vitest';
import { LocalUnansweredStore } from '../adapter/localUnansweredStore';
import { type UnansweredState, UnansweredQueries, type UnansweredStore } from './unansweredQueries';

const T0 = Date.UTC(2026, 10, 16);

class MemoryStore implements UnansweredStore {
  saved: UnansweredState = { enabled: false, entries: [] };
  load() {
    return this.saved;
  }
  save(s: UnansweredState) {
    this.saved = s;
  }
}

describe('UnansweredQueries (#481)', () => {
  it('is off by default and records nothing then', () => {
    const q = new UnansweredQueries(new MemoryStore());
    q.record('zug nach köln');
    expect(q.enabled).toBe(false);
    expect(q.size).toBe(0);
  });

  it('records when switched on and deletes everything when switched off', () => {
    const store = new MemoryStore();
    const q = new UnansweredQueries(store, () => T0);
    q.setEnabled(true);
    q.record('zug nach köln');
    q.record('!g cats');
    expect(q.groups().map((g) => g.word)).toEqual(['zug']);
    q.setEnabled(false);
    expect(store.saved).toEqual({ enabled: false, entries: [] });
  });

  it('keeps the list only in the browser storage and removes it when off', () => {
    const data = new Map<string, string>();
    const storage = {
      getItem: (k: string) => data.get(k) ?? null,
      setItem: (k: string, v: string) => void data.set(k, v),
      removeItem: (k: string) => void data.delete(k),
    } as Storage;
    const store = new LocalUnansweredStore(() => storage);
    expect(store.load()).toEqual({ enabled: false, entries: [] });
    const q = new UnansweredQueries(store, () => T0);
    q.setEnabled(true);
    q.record('pegel rhein');
    expect(new LocalUnansweredStore(() => storage).load().entries).toHaveLength(1);
    q.setEnabled(false);
    expect(data.size).toBe(0);
  });
});
