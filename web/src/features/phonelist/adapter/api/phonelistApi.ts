import type { HttpClient } from '../../../../shared/api-client/http';
import type { EntryDTO } from '../../../../shared/api-client/generated/phonelist';
import type { EntryInputDTO } from '../../../../shared/api-client/generated/phonelist-requests';
import type { PhoneEntry, PhoneEntryInput } from '../../domain/phonelist';
import type { PhonelistGateway } from '../../usecase/phonelistUseCases';

const toEntry = (d: EntryDTO): PhoneEntry => ({
  id: d.id,
  name: d.name,
  role: d.role,
  phone: d.phone,
  dial: d.dial,
  email: d.email,
  team: d.team,
});
const path = (id: string) => `/api/phonelist/entries/${encodeURIComponent(id)}`;

/** PhonelistGateway backed by the REST API. */
export class PhonelistApi implements PhonelistGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<PhoneEntry[]> {
    return (await this.http.get<EntryDTO[]>('/api/phonelist/entries')).map(toEntry);
  }

  async add(input: PhoneEntryInput): Promise<PhoneEntry> {
    const body: EntryInputDTO = { ...input };
    return toEntry(await this.http.post<EntryDTO>('/api/phonelist/entries', body));
  }

  async update(id: string, input: PhoneEntryInput): Promise<PhoneEntry> {
    const body: EntryInputDTO = { ...input };
    return toEntry(await this.http.put<EntryDTO>(path(id), body));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }
}
