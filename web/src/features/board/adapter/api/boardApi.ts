import type { HttpClient } from '../../../../shared/api-client/http';
import type { NoteDTO } from '../../../../shared/api-client/generated/board';
import type { NoteInputDTO } from '../../../../shared/api-client/generated/board-requests';
import { fetchShareGroups, type ShareGroup, type Visibility } from '../../../../shared/sharing/sharing';
import type { BoardNote, BoardNoteInput } from '../../domain/board';
import type { BoardGateway } from '../../usecase/boardUseCases';

const toNote = (d: NoteDTO): BoardNote => ({
  id: d.id,
  text: d.text,
  pinned: d.pinned,
  expiresOn: d.expiresOn,
  visibility: d.visibility,
  mine: d.mine,
});

const toBody = (i: BoardNoteInput, visibility?: Visibility): NoteInputDTO => ({ ...i, visibility });
const path = (id: string) => `/api/board/${encodeURIComponent(id)}`;

/** BoardGateway backed by the REST API. */
export class BoardApi implements BoardGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<BoardNote[]> {
    return (await this.http.get<NoteDTO[]>('/api/board')).map(toNote);
  }

  async create(input: BoardNoteInput, visibility?: Visibility): Promise<BoardNote> {
    return toNote(await this.http.post<NoteDTO>('/api/board', toBody(input, visibility)));
  }

  async update(id: string, input: BoardNoteInput, visibility?: Visibility): Promise<BoardNote> {
    return toNote(await this.http.put<NoteDTO>(path(id), toBody(input, visibility)));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }

  shareGroups(): Promise<ShareGroup[]> {
    return fetchShareGroups(this.http);
  }
}
