import type { DocumentStore } from '../../shared/local/documentStore';
import { exportFileName, exportLocalData } from './localExport';

/**
 * Saves the local data as an export file (ADR 0019, moving to a server). The
 * extension's options page links to the app with #export; there is no
 * server to send a download, so the file is built here and handed to the
 * browser through a temporary object URL.
 */
export async function downloadExport(store: DocumentStore, now: Date = new Date()): Promise<void> {
  const file = await exportLocalData(store, now);
  const url = URL.createObjectURL(new Blob([JSON.stringify(file, null, 2)], { type: 'application/json' }));
  try {
    const a = document.createElement('a');
    a.href = url;
    a.download = exportFileName(now);
    a.click();
  } finally {
    // Revoked on the next tick: the click has started the download by then.
    setTimeout(() => URL.revokeObjectURL(url), 0);
  }
}
