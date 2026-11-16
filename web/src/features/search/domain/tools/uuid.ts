import type { ToolEnv } from './detect';

/** A new UUID (#477). */
export interface UuidResult {
  kind: 'uuid';
  value: string;
  version: 4 | 7;
}

/**
 * "uuid" or "uuid v4" gives a random UUID, "uuid v7" one that starts with
 * the current time in milliseconds, so v7 IDs sort by creation time
 * (RFC 9562). Both are built from env.random, which works over plain HTTP,
 * unlike crypto.randomUUID (#310).
 */
export function uuid(query: string, env: ToolEnv): UuidResult | null {
  const version = /7\s*$/.test(query) ? 7 : 4;
  const b = env.random(16);
  if (version === 7) {
    const ms = env.now().getTime();
    for (let i = 0; i < 6; i++) b[i] = Math.floor(ms / 2 ** (8 * (5 - i))) & 0xff;
  }
  b[6] = ((b[6] ?? 0) & 0x0f) | (version << 4);
  b[8] = ((b[8] ?? 0) & 0x3f) | 0x80; // RFC 9562 variant
  const h = [...b].map((x) => x.toString(16).padStart(2, '0')).join('');
  const value = `${h.slice(0, 8)}-${h.slice(8, 12)}-${h.slice(12, 16)}-${h.slice(16, 20)}-${h.slice(20)}`;
  return { kind: 'uuid', value, version };
}
