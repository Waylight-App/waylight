import { describe, expect, it } from 'vitest';
import { describeConfigResult, describeResult, parseExportFile } from './transfer';

describe('transfer domain', () => {
  it('validates export files', () => {
    expect(parseExportFile('nope')).toMatch(/JSON/);
    expect(parseExportFile('{"format":"other"}')).toMatch(/not a Waylight/);
    expect(parseExportFile('{"format":"waylight-export","version":1,"sections":{}}')).toMatchObject({
      version: 1,
    });
  });
  it('summarises results', () => {
    expect(describeResult({ imported: ['notes'], skipped: ['calendar'], failed: { todos: 'broken' } })).toBe(
      'Imported: notes. Skipped (module disabled or unknown): calendar. Failed: todos (broken).',
    );
  });
  it('summarises an applied config file', () => {
    expect(describeConfigResult({ applied: ['services'], failed: { dashboard: 'unknown widget' } })).toBe(
      'Imported: services. Failed: dashboard (unknown widget).',
    );
  });
});
