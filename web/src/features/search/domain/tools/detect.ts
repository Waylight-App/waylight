/**
 * Small tools in the search field (#477). This file only decides which tool
 * a query is meant for; it stays tiny because it runs on every keystroke and
 * ships with the search. The tools themselves are separate modules that the
 * search imports once a query matches, so none of them adds to the initial
 * bundle. A tool may still find the query invalid and return null; the
 * search then shows its normal results only.
 */

export type ToolName = 'password' | 'uuid' | 'colour' | 'timestamp' | 'subnet';

/** What the tools need from the outside, injected so they stay testable. */
export interface ToolEnv {
  /** n cryptographically random bytes (crypto.getRandomValues). */
  random(n: number): Uint8Array;
  now(): Date;
}

const IPV4 = String.raw`\d{1,3}(?:\.\d{1,3}){3}`;

const patterns: [ToolName, RegExp][] = [
  ['password', /^(?:passwor[dt]|pw|pwd)(?:\s|$)/i],
  ['uuid', /^(?:uuid|guid)(?:\s+v?[47])?$/i],
  ['colour', /^(?:#[0-9a-f]{3,8}|(?:rgba?|hsla?)\s*\(?\s*\d.*)$/i],
  ['timestamp', /^(?:\d{9,13}|unix(?:\s+(?:now|jetzt|time|zeit))?|.+\s+(?:to|in|als|nach)\s+unix)$/i],
  [
    'subnet',
    new RegExp(String.raw`^(?:${IPV4}(?:\s*\/\s*\d{1,2}|\s+${IPV4})|[0-9a-f:]*:[0-9a-f:]*\/\d{1,3})$`, 'i'),
  ],
];

/** The tool a query is meant for, or null. Runs synchronously and cheaply. */
export function detectTool(query: string): ToolName | null {
  const q = query.trim();
  if (!q || q.length > 100) return null;
  return patterns.find(([, re]) => re.test(q))?.[0] ?? null;
}
