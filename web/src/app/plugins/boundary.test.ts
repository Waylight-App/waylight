import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

/** All source files below dir. */
function files(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    return statSync(path).isDirectory() ? files(path) : /\.(ts|svelte)$/.test(name) ? [path] : [];
  });
}

describe('ee/ boundary (ADR 0021)', () => {
  it('no file of the web app imports from ee/', () => {
    const offenders = files('src').filter((f) => /from\s+['"][^'"]*\/ee\//.test(readFileSync(f, 'utf8')));
    expect(offenders).toEqual([]);
  });
});
