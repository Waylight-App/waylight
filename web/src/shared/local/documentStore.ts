/**
 * Storage of the browser extension's local mode (ADR 0019). Each feature
 * keeps one document in the shape of its server tables (rows with IDs and
 * timestamps), so the export in step 4 maps it to the transfer format
 * directly. Local gateways depend on the DocumentStore port only; the
 * IndexedDB implementation is chosen in the local container.
 */
export interface DocumentStore {
  /** The stored document of a feature, or undefined before its first save. */
  get<T>(feature: string): Promise<T | undefined>;
  put<T>(feature: string, doc: T): Promise<void>;
}

/** Keeps documents in memory; for tests and as fallback without IndexedDB. */
export class MemoryDocumentStore implements DocumentStore {
  private readonly docs = new Map<string, string>();

  get<T>(feature: string): Promise<T | undefined> {
    const json = this.docs.get(feature);
    return Promise.resolve(json === undefined ? undefined : (JSON.parse(json) as T));
  }

  put<T>(feature: string, doc: T): Promise<void> {
    // Stored as JSON so callers can never mutate what is saved by accident.
    this.docs.set(feature, JSON.stringify(doc));
    return Promise.resolve();
  }
}

const DB = 'trisse-local';
const STORE = 'documents';

/**
 * One object store, keyed by feature name. A private window without
 * IndexedDB falls back to memory: the page works, but nothing survives
 * closing it.
 */
export class IndexedDbDocumentStore implements DocumentStore {
  private readonly fallback = new MemoryDocumentStore();
  private db: Promise<IDBDatabase> | undefined;

  private open(): Promise<IDBDatabase> {
    this.db ??= new Promise((resolve, reject) => {
      const req = indexedDB.open(DB, 1);
      req.onupgradeneeded = () => req.result.createObjectStore(STORE);
      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error ?? new Error('IndexedDB unavailable'));
    });
    return this.db;
  }

  private async run<T>(mode: IDBTransactionMode, fn: (s: IDBObjectStore) => IDBRequest): Promise<T> {
    const db = await this.open();
    return new Promise<T>((resolve, reject) => {
      const tx = db.transaction(STORE, mode);
      const req = fn(tx.objectStore(STORE));
      tx.oncomplete = () => resolve(req.result as T);
      tx.onerror = () => reject(tx.error ?? new Error('IndexedDB failed'));
    });
  }

  async get<T>(feature: string): Promise<T | undefined> {
    if (typeof indexedDB === 'undefined') return this.fallback.get<T>(feature);
    return this.run<T | undefined>('readonly', (s) => s.get(feature));
  }

  async put<T>(feature: string, doc: T): Promise<void> {
    if (typeof indexedDB === 'undefined') return this.fallback.put(feature, doc);
    // Structured clone rejects Svelte state proxies; a JSON round trip makes
    // the document plain.
    const plain = JSON.parse(JSON.stringify(doc)) as T;
    await this.run('readwrite', (s) => s.put(plain, feature));
  }
}
