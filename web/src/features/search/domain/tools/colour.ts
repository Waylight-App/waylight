/** One colour in the notations people paste around (#477). */
export interface ColourResult {
  kind: 'colour';
  /** #rrggbb, or #rrggbbaa when not fully opaque. */
  hex: string;
  rgb: string;
  hsl: string;
  oklch: string;
  /** CSS value for a swatch; the same colour as hex. */
  css: string;
}

interface Rgba {
  r: number; // 0–255
  g: number;
  b: number;
  a: number; // 0–1
}

/**
 * Reads "#1e90ff", "#09f", "rgb(30,144,255)", "rgba(30 144 255 / 50%)",
 * "hsl(210, 100%, 56%)" or "hsl 210 100 56" and shows the same colour in
 * hex, RGB, HSL and OKLCH. Anything out of range yields null.
 */
export function colour(query: string): ColourResult | null {
  const c = parse(query.trim());
  return c ? describe(c) : null;
}

function parse(q: string): Rgba | null {
  const hex = /^#([0-9a-f]{3,4}|[0-9a-f]{6}|[0-9a-f]{8})$/i.exec(q)?.[1];
  if (hex) {
    const full = hex.length <= 4 ? [...hex].map((d) => d + d).join('') : hex;
    const n = (i: number) => parseInt(full.slice(i, i + 2), 16);
    return { r: n(0), g: n(2), b: n(4), a: full.length === 8 ? n(6) / 255 : 1 };
  }
  const fn = /^(rgba?|hsla?)\s*\(?\s*(.*?)\s*\)?$/i.exec(q);
  if (!fn) return null;
  const parts = (fn[2] ?? '').split(/[\s,/]+/).filter(Boolean);
  if (parts.length < 3 || parts.length > 4) return null;
  const alpha = parts[3] === undefined ? 1 : number(parts[3], 1);
  if (alpha === null || alpha < 0 || alpha > 1) return null;
  if ((fn[1] ?? '').toLowerCase().startsWith('rgb')) {
    const [r, g, b] = parts.slice(0, 3).map((p) => number(p, 255));
    if (![r, g, b].every((v) => v !== null && v !== undefined && v >= 0 && v <= 255)) return null;
    return { r: Math.round(r as number), g: Math.round(g as number), b: Math.round(b as number), a: alpha };
  }
  const h = Number.parseFloat((parts[0] ?? '').replace(/deg$/i, ''));
  const s = number(parts[1] ?? '', 100);
  const l = number(parts[2] ?? '', 100);
  if (!Number.isFinite(h) || s === null || l === null || s < 0 || s > 100 || l < 0 || l > 100) return null;
  return { ...hslToRgb(h, s / 100, l / 100), a: alpha };
}

/** A number; "50%" is read as a share of max. */
function number(text: string, max: number): number | null {
  const pct = text.endsWith('%');
  const n = Number(pct ? text.slice(0, -1) : text);
  if (text === '' || !Number.isFinite(n)) return null;
  return pct ? (n / 100) * max : n;
}

function hslToRgb(h: number, s: number, l: number): { r: number; g: number; b: number } {
  const k = (n: number) => (n + (((h % 360) + 360) % 360) / 30) % 12;
  const a = s * Math.min(l, 1 - l);
  const f = (n: number) => l - a * Math.max(-1, Math.min(k(n) - 3, 9 - k(n), 1));
  return { r: Math.round(f(0) * 255), g: Math.round(f(8) * 255), b: Math.round(f(4) * 255) };
}

function rgbToHsl({ r, g, b }: Rgba): { h: number; s: number; l: number } {
  const [rn, gn, bn] = [r / 255, g / 255, b / 255];
  const max = Math.max(rn, gn, bn);
  const min = Math.min(rn, gn, bn);
  const l = (max + min) / 2;
  const d = max - min;
  if (d === 0) return { h: 0, s: 0, l };
  const s = d / (1 - Math.abs(2 * l - 1));
  const h = max === rn ? ((gn - bn) / d) % 6 : max === gn ? (bn - rn) / d + 2 : (rn - gn) / d + 4;
  return { h: (h * 60 + 360) % 360, s, l };
}

/** sRGB → OKLab → OKLCH (Björn Ottosson's reference matrices). */
function rgbToOklch({ r, g, b }: Rgba): { l: number; c: number; h: number } {
  const lin = (v: number) => {
    const x = v / 255;
    return x <= 0.04045 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4;
  };
  const [lr, lg, lb] = [lin(r), lin(g), lin(b)];
  const l_ = Math.cbrt(0.4122214708 * lr + 0.5363325363 * lg + 0.0514459929 * lb);
  const m_ = Math.cbrt(0.2119034982 * lr + 0.6806995451 * lg + 0.1073969566 * lb);
  const s_ = Math.cbrt(0.0883024619 * lr + 0.2817188376 * lg + 0.6299787005 * lb);
  const L = 0.2104542553 * l_ + 0.793617785 * m_ - 0.0040720468 * s_;
  const A = 1.9779984951 * l_ - 2.428592205 * m_ + 0.4505937099 * s_;
  const B = 0.0259040371 * l_ + 0.7827717662 * m_ - 0.808675766 * s_;
  const c = Math.hypot(A, B);
  return { l: L, c, h: c < 1e-4 ? 0 : ((Math.atan2(B, A) * 180) / Math.PI + 360) % 360 };
}

const round = (n: number, digits = 0) => Number(n.toFixed(digits));
const hex2 = (n: number) => n.toString(16).padStart(2, '0');

function describe(c: Rgba): ColourResult {
  const opaque = c.a >= 1;
  const alpha = round(c.a, 2);
  const hex = `#${hex2(c.r)}${hex2(c.g)}${hex2(c.b)}${opaque ? '' : hex2(Math.round(c.a * 255))}`;
  const { h, s, l } = rgbToHsl(c);
  const ok = rgbToOklch(c);
  const slash = opaque ? '' : ` / ${alpha}`;
  return {
    kind: 'colour',
    hex,
    rgb: `rgb(${c.r} ${c.g} ${c.b}${slash})`,
    hsl: `hsl(${round(h)} ${round(s * 100)}% ${round(l * 100)}%${slash})`,
    oklch: `oklch(${round(ok.l * 100, 1)}% ${round(ok.c, 3)} ${round(ok.h, 1)}${slash})`,
    css: hex,
  };
}
