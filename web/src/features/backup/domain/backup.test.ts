import { describe, expect, it } from 'vitest';
import { downloadUrl } from './backup';

describe('downloadUrl', () => {
  it('encodes the file name', () => {
    expect(downloadUrl({ name: 'waylight-20261111-030000.db', size: 1, createdAt: new Date() })).toBe(
      '/api/backups/waylight-20261111-030000.db',
    );
  });
});
