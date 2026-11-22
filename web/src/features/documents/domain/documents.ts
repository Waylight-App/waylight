import type { Visibility } from '../../../shared/sharing/sharing';

/** Warranty state of a document on the server's day. */
export const STATUSES = ['none', 'running', 'ending', 'ended'] as const;
export type Status = (typeof STATUSES)[number];

/** Limits of an attached file; the server checks them again. */
export const MAX_FILE_BYTES = 10 * 1024 * 1024;
export const FILE_ACCEPT = 'image/jpeg,image/png,image/webp,application/pdf';
const FILE_TYPES = FILE_ACCEPT.split(',');

/** A product with its warranty and an optional file (#688). */
export interface Doc {
  id: string;
  title: string;
  /** YYYY-MM-DD or empty. */
  purchasedOn: string;
  warrantyEnd: string;
  note: string;
  status: Status;
  /** Empty when no file is attached. */
  fileName: string;
  contentType: string;
  fileSize: number;
  visibility: Visibility;
  mine: boolean;
}

/** What the dialog edits. */
export interface DocInput {
  title: string;
  purchasedOn: string;
  warrantyEnd: string;
  note: string;
}

/** Message keys of validateDoc and validateFile. */
export type DocError =
  | 'documents.error.title'
  | 'documents.error.dates'
  | 'documents.error.file_type'
  | 'documents.error.file_size';

/** Validates the dialog and returns a message key or null. */
export function validateDoc(input: DocInput): DocError | null {
  const title = input.title.trim();
  if (!title || [...title].length > 80) return 'documents.error.title';
  if (input.purchasedOn && input.warrantyEnd && input.warrantyEnd < input.purchasedOn)
    return 'documents.error.dates';
  return null;
}

/** Checks a chosen file before it is sent: a photo or a PDF of up to 10 MB. */
export function validateFile(file: { type: string; size: number }): DocError | null {
  if (!FILE_TYPES.includes(file.type)) return 'documents.error.file_type';
  if (file.size === 0 || file.size > MAX_FILE_BYTES) return 'documents.error.file_size';
  return null;
}

/** File size for the list, e.g. "2.4 MB". */
export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
