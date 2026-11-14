import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError, FetchHttpClient } from './http';

const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

describe('FetchHttpClient GET sharing (#318)', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('shares identical GETs in flight and copies the answer per caller', async () => {
    const fetchMock = vi.fn().mockImplementation(async () => json({ points: [1, 2] }));
    vi.stubGlobal('fetch', fetchMock);
    const http = new FetchHttpClient(() => {});
    const [a, b] = await Promise.all([
      http.get<{ points: number[] }>('/api/system/history?hours=24'),
      http.get<{ points: number[] }>('/api/system/history?hours=24'),
    ]);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    a.points.push(3);
    expect(b.points).toEqual([1, 2]);
    // Afterwards a new request is made again (no caching beyond the flight).
    await http.get('/api/system/history?hours=24');
    expect(fetchMock).toHaveBeenCalledTimes(2);
    // Different paths are separate.
    await Promise.all([http.get('/api/a'), http.get('/api/b')]);
    expect(fetchMock).toHaveBeenCalledTimes(4);
  });

  it('gives every caller the error of a shared failed GET', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation(async () => json({ error: { code: 'not_found', message: 'nope' } }, 404)),
    );
    const http = new FetchHttpClient(() => {});
    const results = await Promise.allSettled([http.get('/api/x'), http.get('/api/x')]);
    for (const r of results)
      expect(r.status === 'rejected' && r.reason instanceof ApiError && r.reason.status).toBe(404);
  });
});
