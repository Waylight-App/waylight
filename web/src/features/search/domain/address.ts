/**
 * Detects input that is clearly an address (#475), so Enter opens it
 * instead of searching for it: "github.com", "192.168.178.1",
 * "nas.local:5001", "http://router". Anything with a space, a version
 * number or a plain decimal stays a search.
 */

/**
 * Top-level domains that turn "name.tld" into an address. A compact list of
 * the common generic and country domains; rarer ones still work with an
 * explicit "https://". "js" is deliberately missing: "node.js" is a search.
 */
const TLDS = new Set(
  (
    'com org net edu gov mil int io co ai app dev info biz me tv cc xyz online site tech store blog cloud ' +
    'page wiki news shop live pro eu de at ch nl be fr it es pt pl cz sk hu dk se no fi is ie uk gb us ca ' +
    'mx br ar cl au nz jp cn kr in sg hk tw ru ua tr gr ro bg hr si rs lt lv ee lu li za il ae'
  ).split(' '),
);

/** Local names that are addresses without a public top-level domain. */
const LOCAL_SUFFIXES = ['.local', '.lan', '.home.arpa', '.fritz.box', '.internal'];

const IPV4 = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/;
const PORT_PATH = /^(?::\d{1,5})?(?:[/?#].*)?$/;
const LABEL = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/;

function isIPv4(host: string): boolean {
  const m = IPV4.exec(host);
  return !!m && m.slice(1).every((part) => Number(part) <= 255);
}

/** Splits "host:port/path" into the host and the rest. */
function splitHost(text: string): { host: string; rest: string } {
  const i = text.search(/[:/?#]/);
  return i < 0 ? { host: text, rest: '' } : { host: text.slice(0, i), rest: text.slice(i) };
}

/**
 * The URL to open for the typed text, or null when it is not clearly an
 * address. The scheme is added when missing: http:// for IP addresses and
 * local names (home servers rarely have certificates), https:// otherwise.
 */
export function detectAddress(input: string): string | null {
  const text = input.trim();
  if (text === '' || /\s/.test(text)) return null;
  if (/^https?:\/\/[^\s/?#]+/i.test(text)) {
    try {
      return new URL(text).href;
    } catch {
      return null;
    }
  }
  const { host: rawHost, rest } = splitHost(text);
  if (!PORT_PATH.test(rest)) return null;
  const host = rawHost.toLowerCase();
  const local =
    isIPv4(host) ||
    host === 'localhost' ||
    host === 'fritz.box' || // the FRITZ!Box answers to its bare name
    LOCAL_SUFFIXES.some((s) => host.endsWith(s) && host.length > s.length);
  if (!local) {
    const labels = host.split('.');
    if (labels.length < 2 || !labels.every((l) => LABEL.test(l))) return null;
    const tld = labels[labels.length - 1] ?? '';
    // "v1.2.3" and "5.5" end in digits; only known domains count.
    if (!TLDS.has(tld)) return null;
  }
  const url = `${local ? 'http' : 'https'}://${text}`;
  try {
    return new URL(url).href;
  } catch {
    return null;
  }
}

/** The address as shown on the row: without scheme and a trailing slash. */
export function displayAddress(url: string): string {
  return url.replace(/^https?:\/\//, '').replace(/\/$/, '');
}
