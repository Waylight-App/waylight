import type { HttpClient } from '../../../../shared/api-client/http';
import type { DocumentDTO } from '../../../../shared/api-client/generated/documents';
import type { DocumentInputDTO } from '../../../../shared/api-client/generated/documents-requests';
import { fetchShareGroups, type ShareGroup, type Visibility } from '../../../../shared/sharing/sharing';
import { type Doc, STATUSES, type Status } from '../../domain/documents';
import type { DocData, DocumentsGateway } from '../../usecase/documentsUseCases';

const toStatus = (s: string): Status =>
  (STATUSES as readonly string[]).includes(s) ? (s as Status) : 'none';

const toDoc = (d: DocumentDTO): Doc => ({ ...d, status: toStatus(d.status) });
const toBody = (d: DocData, visibility?: Visibility): DocumentInputDTO => ({ ...d, visibility });
const path = (id: string) => `/api/documents/${encodeURIComponent(id)}`;

/** DocumentsGateway backed by the REST API. */
export class DocumentsApi implements DocumentsGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<Doc[]> {
    return (await this.http.get<DocumentDTO[]>('/api/documents')).map(toDoc);
  }

  async create(data: DocData, visibility?: Visibility): Promise<Doc> {
    return toDoc(await this.http.post<DocumentDTO>('/api/documents', toBody(data, visibility)));
  }

  async update(id: string, data: DocData, visibility?: Visibility): Promise<Doc> {
    return toDoc(await this.http.put<DocumentDTO>(path(id), toBody(data, visibility)));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }

  async attachFile(id: string, file: File): Promise<Doc> {
    const form = new FormData();
    form.append('file', file, file.name);
    return toDoc(await this.http.postForm<DocumentDTO>(`${path(id)}/file`, form));
  }

  fileUrl(id: string): string {
    return `${path(id)}/file`;
  }

  shareGroups(): Promise<ShareGroup[]> {
    return fetchShareGroups(this.http);
  }
}
