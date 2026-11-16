import { describe, expect, it } from 'vitest';
import { colour } from './colour';
import { detectTool, type ToolEnv } from './detect';
import { generate, password } from './password';
import { runTool } from './run';
import { subnet } from './subnet';
import { parseDate, timestamp } from './timestamp';
import { uuid } from './uuid';

/** Deterministic bytes, so the tests do not depend on chance. */
function fakeEnv(now = new Date('2026-11-16T12:00:00Z')): ToolEnv {
  let seed = 7;
  return {
    random: (n) => Uint8Array.from({ length: n }, () => (seed = (seed * 73 + 41) % 256)),
    now: () => now,
  };
}
const env: ToolEnv = {
  random: (n) => crypto.getRandomValues(new Uint8Array(n)),
  now: () => new Date('2026-11-16T12:00:00Z'),
};

describe('detectTool', () => {
  it.each([
    ['passwort', 'password'],
    ['pw 32 ohne sonderzeichen', 'password'],
    ['uuid v7', 'uuid'],
    ['#1e90ff', 'colour'],
    ['rgb(30,144,255)', 'colour'],
    ['hsl 210 100 56', 'colour'],
    ['1700000000', 'timestamp'],
    ['unix now', 'timestamp'],
    ['2026-11-16 12:00 to unix', 'timestamp'],
    ['192.168.178.0/24', 'subnet'],
    ['10.0.0.5 255.255.255.0', 'subnet'],
    ['2001:db8::/48', 'subnet'],
  ])('%s → %s', (q, tool) => expect(detectTool(q)).toBe(tool));

  it.each([
    'weather',
    'passwords manager',
    'uuidgen tips',
    '#hashtag',
    '12345',
    '2+3',
    'unix philosophy book',
  ])('leaves %s to the normal search', (q) => expect(detectTool(q)).toBeNull());
});

describe('password', () => {
  it('uses the default length and every character set', () => {
    const p = password('passwort', env);
    expect(p?.value).toHaveLength(20);
    expect(p?.value).toMatch(/[a-z]/);
    expect(p?.value).toMatch(/[A-Z]/);
    expect(p?.value).toMatch(/[2-9]/);
    expect(p?.value).toMatch(/[!#$%&*+\-=?@_~]/);
  });
  it('reads length and options', () => {
    const p = password('pw 32 ohne sonderzeichen', env);
    expect(p?.value).toMatch(/^[A-Za-z2-9]{32}$/);
    expect(password('pin 6', env)?.value).toMatch(/^\d{6}$/);
  });
  it('falls through for impossible lengths', () => {
    expect(password('password 2', env)).toBeNull();
    expect(password('password 500', env)).toBeNull();
    expect(password('password 12 24', env)).toBeNull();
  });
  it('puts one character of each set even when chance would not', () => {
    // Bytes that always pick the first character of the combined set.
    const zeros: ToolEnv = { random: (n) => new Uint8Array(n), now: () => new Date() };
    expect(generate(4, ['ab', 'CD', '23'], zeros)).toMatch(/(?=.*[ab])(?=.*[CD])(?=.*[23])/);
  });
  it('never repeats over many draws (sanity check of the randomness)', () => {
    const seen = new Set(Array.from({ length: 200 }, () => password('pw 16', env)?.value));
    expect(seen.size).toBe(200);
  });
});

describe('uuid', () => {
  it('builds RFC 9562 v4 and v7 UUIDs', () => {
    const v4 = uuid('uuid', fakeEnv())?.value;
    expect(v4).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
    const v7 = uuid('uuid v7', fakeEnv())?.value ?? '';
    expect(v7).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab]/);
    // The first 48 bits are the time in milliseconds.
    expect(parseInt(v7.replace(/-/g, '').slice(0, 12), 16)).toBe(Date.parse('2026-11-16T12:00:00Z'));
  });
});

describe('colour', () => {
  it('converts between notations', () => {
    const c = colour('#1e90ff');
    expect(c).toMatchObject({ hex: '#1e90ff', rgb: 'rgb(30 144 255)', hsl: 'hsl(210 100% 56%)' });
    expect(c?.oklch).toBe('oklch(65.2% 0.19 253.2)');
    expect(colour('rgb(30,144,255)')?.hex).toBe('#1e90ff');
    expect(colour('hsl 210 100 56')?.hex).toBe('#1f8fff');
    expect(colour('#09f')?.hex).toBe('#0099ff');
  });
  it('keeps transparency', () => {
    expect(colour('rgba(255 0 0 / 50%)')).toMatchObject({ hex: '#ff000080', rgb: 'rgb(255 0 0 / 0.5)' });
  });
  it('rejects values out of range', () => {
    expect(colour('rgb(300, 0, 0)')).toBeNull();
    expect(colour('#12345')).toBeNull();
    expect(colour('hsl(10, 120%, 50%)')).toBeNull();
  });
  it('gives white and black their exact OKLCH', () => {
    expect(colour('#ffffff')?.oklch).toBe('oklch(100% 0 0)');
    expect(colour('#000000')?.oklch).toBe('oklch(0% 0 0)');
  });
});

describe('timestamp', () => {
  it('reads seconds and milliseconds', () => {
    expect(timestamp('1700000000', env)?.date.toISOString()).toBe('2023-11-14T22:13:20.000Z');
    expect(timestamp('1700000000123', env)?.seconds).toBe(1700000000);
  });
  it('gives the current timestamp', () => {
    expect(timestamp('unix now', env)?.seconds).toBe(Date.parse('2026-11-16T12:00:00Z') / 1000);
  });
  it('converts a date to a timestamp', () => {
    expect(timestamp('2026-11-16 12:00 UTC to unix', env)?.seconds).toBe(1794830400);
    expect(timestamp('16.11.2026 12:00 in unix', env)?.date.getHours()).toBe(12);
  });
  it('rejects impossible dates', () => {
    expect(parseDate('2026-02-31')).toBeNull();
    expect(timestamp('next tuesday to unix', env)).toBeNull();
  });
});

describe('subnet', () => {
  it('describes an IPv4 network', () => {
    expect(subnet('192.168.178.17/24')).toMatchObject({
      network: '192.168.178.0/24',
      mask: '255.255.255.0',
      broadcast: '192.168.178.255',
      firstHost: '192.168.178.1',
      lastHost: '192.168.178.254',
      hosts: '254',
    });
    expect(subnet('10.0.0.5 255.255.255.0')?.network).toBe('10.0.0.0/24');
  });
  it('handles point-to-point and host routes', () => {
    expect(subnet('10.0.0.0/31')).toMatchObject({ hosts: '2', broadcast: undefined, lastHost: '10.0.0.1' });
    expect(subnet('10.0.0.7/32')).toMatchObject({ hosts: '1', firstHost: '10.0.0.7' });
  });
  it('describes an IPv6 prefix', () => {
    expect(subnet('2001:db8:0:1::5/64')).toMatchObject({
      network: '2001:db8:0:1::/64',
      firstHost: '2001:db8:0:1::',
      lastHost: '2001:db8:0:1:ffff:ffff:ffff:ffff',
      hosts: '18446744073709551616',
    });
  });
  it('rejects invalid input', () => {
    expect(subnet('300.1.1.1/24')).toBeNull();
    expect(subnet('10.0.0.1 255.0.255.0')).toBeNull();
    expect(subnet('10.0.0.1/33')).toBeNull();
    expect(subnet('1::2::3/64')).toBeNull();
  });
});

describe('runTool', () => {
  it('loads a tool and lets invalid input fall through', async () => {
    expect((await runTool('subnet', '10.0.0.0/8', env))?.kind).toBe('subnet');
    expect(await runTool('colour', 'rgb(1,2)', env)).toBeNull();
  });
});
