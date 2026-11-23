import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

/** All .svelte and .css files below dir. */
function sources(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) return sources(p);
    return /\.(svelte|css)$/.test(name) ? [p] : [];
  });
}

describe('design tokens (#113)', () => {
  it('every color token in use is defined, so no fallback with poor contrast kicks in', () => {
    const src = join(__dirname, '..', '..');
    const defined = new Set(
      readFileSync(join(__dirname, 'tokens.css'), 'utf8').match(/--color-[a-z0-9-]+(?=\s*:)/g),
    );
    const undefinedTokens = sources(src).flatMap((file) =>
      [...readFileSync(file, 'utf8').matchAll(/var\((--color-[a-z0-9-]+)/g)]
        .map((m) => m[1] ?? '')
        .filter((token) => !defined.has(token))
        .map((token) => `${file.slice(src.length + 1)}: ${token}`),
    );
    expect(undefinedTokens).toEqual([]);
  });
});

/** WCAG relative luminance of #rgb or #rrggbb. */
function luminance(hex: string): number {
  const h = hex.length === 4 ? [...hex.slice(1)].map((c) => c + c).join('') : hex.slice(1);
  const [r = 0, g = 0, b = 0] = [0, 2, 4].map((i) => {
    const c = parseInt(h.slice(i, i + 2), 16) / 255;
    return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x) as [number, number];
  return (hi + 0.05) / (lo + 0.05);
}

/** Colour tokens of every rule in tokens.css, by selector. */
function blocks(css: string): Map<string, Record<string, string>> {
  const out = new Map<string, Record<string, string>>();
  const plain = css.replace(/\/\*[\s\S]*?\*\//g, '');
  for (const [, selector = '', body = ''] of plain.matchAll(/([^{}]+)\{([^}]*)\}/g)) {
    out.set(
      selector.trim(),
      Object.fromEntries(
        [...body.matchAll(/(--color-[a-z0-9-]+):\s*(#[0-9a-f]{3,6})\b/gi)].map((m) => [m[1], m[2]]),
      ),
    );
  }
  return out;
}

const pairs: [string, string[]][] = [
  ['--color-text', ['--color-bg', '--color-surface', '--color-surface-2']],
  ['--color-text-muted', ['--color-bg', '--color-surface', '--color-surface-2']],
  ['--color-accent-text', ['--color-surface', '--color-surface-2']],
  ['--color-danger', ['--color-surface', '--color-surface-2']],
  ['--color-success', ['--color-surface', '--color-surface-2']],
  ['--color-warning', ['--color-surface', '--color-surface-2']],
  ['--color-accent-contrast', ['--color-accent']],
];

function failing(t: Record<string, string>, min: number): string[] {
  return pairs.flatMap(([fg, bgs]) =>
    bgs
      .map((bg) => [bg, contrast(t[fg] ?? '#000', t[bg] ?? '#fff')] as const)
      .filter(([, c]) => c < min)
      .map(([bg, c]) => `${fg} on ${bg}: ${c.toFixed(2)}`),
  );
}

const rules = blocks(readFileSync(join(__dirname, 'tokens.css'), 'utf8'));
const light = rules.get(':root') ?? {};
const dark = { ...light, ...rules.get(":root[data-theme='dark']") };

describe('contrast of the light and dark tokens (#113)', () => {
  const schemes = { light, dark };

  it.each(Object.entries(schemes))('%s meets WCAG AA (4.5:1) for text', (_name, t) => {
    expect(failing(t, 4.5)).toEqual([]);
  });
});

describe('contrast of the high-contrast tokens (#681)', () => {
  const highLight = rules.get(":root:root:root[data-contrast='high']") ?? {};
  const highDark = { ...highLight, ...rules.get(":root:root:root[data-contrast='high'][data-theme='dark']") };

  it('sets every colour the checks need', () => {
    for (const t of [highLight, highDark])
      for (const [fg, bgs] of pairs) for (const name of [fg, ...bgs]) expect(t[name], name).toBeDefined();
  });

  it.each([
    ['light', highLight],
    ['dark', highDark],
  ])('%s meets WCAG AAA (7:1) for text', (_name, t) => {
    expect(failing(t, 7)).toEqual([]);
  });
});
