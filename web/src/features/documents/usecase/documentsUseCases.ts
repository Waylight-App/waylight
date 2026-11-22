import type { ShareGroup, Visibility } from '../../../shared/sharing/sharing';
import type { Doc, DocInput } from '../domain/documents';

/** The fields the backend stores. */
export type DocData = DocInput;

/** Output port: documents backend. */
export interface DocumentsGateway {
  list(): Promise<Doc[]>;
  create(data: DocData, visibility?: Visibility): Promise<Doc>;
  update(id: string, data: DocData, visibility?: Visibility): Promise<Doc>;
  remove(id: string): Promise<void>;
  attachFile(id: string, file: File): Promise<Doc>;
  fileUrl(id: string): string;
  shareGroups(): Promise<ShareGroup[]>;
}

const toData = (i: DocInput): DocData => ({
  title: i.title.trim(),
  purchasedOn: i.purchasedOn,
  warrantyEnd: i.warrantyEnd,
  note: i.note.trim(),
});

/** Application logic of the documents (#688). */
export class DocumentsUseCases {
  constructor(private readonly gateway: DocumentsGateway) {}

  list(): Promise<Doc[]> {
    return this.gateway.list();
  }

  /**
   * Enters (no id) or changes a document and attaches the chosen file. A
   * file the server refuses leaves the entry without it, so the dialog
   * reports the error and the next save sends the file again.
   */
  async save(id: string | null, input: DocInput, file: File | null, visibility?: Visibility): Promise<Doc> {
    const data = toData(input);
    const saved = id
      ? await this.gateway.update(id, data, visibility)
      : await this.gateway.create(data, visibility);
    return file ? this.gateway.attachFile(saved.id, file) : saved;
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  /** Address of the download; the browser sends the session cookie. */
  fileUrl(id: string): string {
    return this.gateway.fileUrl(id);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return this.gateway.shareGroups();
  }
}
