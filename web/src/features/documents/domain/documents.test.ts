import { describe, expect, it } from 'vitest';
import { formatSize, MAX_FILE_BYTES, validateDoc, validateFile } from './documents';

describe('documents domain', () => {
  const ok = { title: 'Kettle', purchasedOn: '2026-01-02', warrantyEnd: '2028-01-02', note: '' };

  it('validates the dialog', () => {
    expect(validateDoc(ok)).toBeNull();
    expect(validateDoc({ ...ok, purchasedOn: '', warrantyEnd: '' })).toBeNull();
    expect(validateDoc({ ...ok, title: '  ' })).toBe('documents.error.title');
    expect(validateDoc({ ...ok, title: 'x'.repeat(81) })).toBe('documents.error.title');
    expect(validateDoc({ ...ok, warrantyEnd: '2025-12-31' })).toBe('documents.error.dates');
  });

  it('accepts photos and PDFs of up to 10 MB only', () => {
    expect(validateFile({ type: 'application/pdf', size: 1000 })).toBeNull();
    expect(validateFile({ type: 'image/webp', size: MAX_FILE_BYTES })).toBeNull();
    expect(validateFile({ type: 'image/gif', size: 1000 })).toBe('documents.error.file_type');
    expect(validateFile({ type: 'text/html', size: 1000 })).toBe('documents.error.file_type');
    expect(validateFile({ type: 'image/png', size: MAX_FILE_BYTES + 1 })).toBe('documents.error.file_size');
    expect(validateFile({ type: 'image/png', size: 0 })).toBe('documents.error.file_size');
  });

  it('formats sizes', () => {
    expect(formatSize(512)).toBe('512 B');
    expect(formatSize(2048)).toBe('2 KB');
    expect(formatSize(2.5 * 1024 * 1024)).toBe('2.5 MB');
  });
});
