/** The facts about one network (#477). */
export interface SubnetResult {
  kind: 'subnet';
  family: 4 | 6;
  /** Network address with prefix, e.g. "192.168.178.0/24". */
  network: string;
  prefix: number;
  /** Dotted mask (IPv4 only). */
  mask?: string;
  /** Broadcast address (IPv4 only, not for /31 and /32). */
  broadcast?: string;
  firstHost: string;
  lastHost: string;
  /** Usable hosts as a decimal string (IPv6 counts do not fit a number). */
  hosts: string;
}

/**
 * "192.168.178.0/24", "10.0.0.5 255.255.255.0" or "2001:db8::/48". For
 * IPv4, /31 links have two usable hosts and /32 one (RFC 3021); for IPv6
 * every address of the prefix counts.
 */
export function subnet(query: string): SubnetResult | null {
  const q = query.trim();
  return q.includes(':') ? ipv6(q) : ipv4(q);
}

function ipv4(q: string): SubnetResult | null {
  const m = /^(\S+?)\s*(?:\/\s*(\d{1,2})|\s+(\S+))$/.exec(q);
  if (!m) return null;
  const addr = parse4(m[1] ?? '');
  if (addr === null) return null;
  let prefix: number;
  if (m[2] !== undefined) prefix = Number(m[2]);
  else {
    const mask = parse4(m[3] ?? '');
    if (mask === null) return null;
    prefix = maskToPrefix(mask);
  }
  if (!(prefix >= 0 && prefix <= 32)) return null;
  const mask = prefix === 0 ? 0 : (0xffffffff << (32 - prefix)) >>> 0;
  const network = (addr & mask) >>> 0;
  const broadcast = (network | ~mask) >>> 0;
  const size = 2 ** (32 - prefix);
  const small = prefix >= 31;
  return {
    kind: 'subnet',
    family: 4,
    network: `${format4(network)}/${prefix}`,
    prefix,
    mask: format4(mask),
    broadcast: small ? undefined : format4(broadcast),
    firstHost: format4(small ? network : network + 1),
    lastHost: format4(small ? broadcast : broadcast - 1),
    hosts: String(small ? size : size - 2),
  };
}

function parse4(text: string): number | null {
  const parts = text.split('.');
  if (parts.length !== 4 || !parts.every((p) => /^\d{1,3}$/.test(p) && Number(p) <= 255)) return null;
  return parts.reduce((n, p) => n * 256 + Number(p), 0);
}

function format4(n: number): string {
  return [24, 16, 8, 0].map((s) => (n >>> s) & 0xff).join('.');
}

/** Prefix length of a mask; NaN when the ones are not contiguous. */
function maskToPrefix(mask: number): number {
  const inverted = ~mask >>> 0;
  if ((inverted & (inverted + 1)) !== 0) return Number.NaN;
  return 32 - Math.log2(inverted + 1);
}

function ipv6(q: string): SubnetResult | null {
  const m = /^([0-9a-f:]+)\/(\d{1,3})$/i.exec(q);
  if (!m) return null;
  const addr = parse6(m[1] ?? '');
  const prefix = Number(m[2]);
  if (addr === null || prefix > 128) return null;
  const all = (1n << 128n) - 1n;
  const hostBits = BigInt(128 - prefix);
  const mask = (all << hostBits) & all;
  const network = addr & mask;
  const last = network | (~mask & all);
  return {
    kind: 'subnet',
    family: 6,
    network: `${format6(network)}/${prefix}`,
    prefix,
    firstHost: format6(network),
    lastHost: format6(last),
    hosts: (1n << hostBits).toString(),
  };
}

function parse6(text: string): bigint | null {
  const halves = text.split('::');
  if (halves.length > 2) return null;
  const groups = (s: string | undefined) => (s ? s.split(':') : []);
  const head = groups(halves[0]);
  const tail = groups(halves[1]);
  const missing = 8 - head.length - tail.length;
  if (halves.length === 1 ? missing !== 0 : missing < 1) return null;
  const all = [...head, ...Array<string>(halves.length === 2 ? missing : 0).fill('0'), ...tail];
  if (!all.every((g) => /^[0-9a-f]{1,4}$/i.test(g))) return null;
  return all.reduce((n, g) => (n << 16n) | BigInt(parseInt(g, 16)), 0n);
}

/** RFC 5952 text form: lower case, longest run of zero groups as "::". */
function format6(n: bigint): string {
  const g = Array.from({ length: 8 }, (_, i) => Number((n >> BigInt(16 * (7 - i))) & 0xffffn));
  let best = { at: -1, len: 1 };
  for (let i = 0; i < 8;) {
    if (g[i] !== 0) {
      i++;
      continue;
    }
    let j = i;
    while (j < 8 && g[j] === 0) j++;
    if (j - i > best.len) best = { at: i, len: j - i };
    i = j;
  }
  const hex = g.map((x) => x.toString(16));
  if (best.at < 0) return hex.join(':');
  return `${hex.slice(0, best.at).join(':')}::${hex.slice(best.at + best.len).join(':')}`;
}
