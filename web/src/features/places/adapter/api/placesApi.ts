import type { HttpClient } from '../../../../shared/api-client/http';
import type { PlaceDTO } from '../../../../shared/api-client/generated/places';
import type { PlaceInputDTO } from '../../../../shared/api-client/generated/places-requests';
import type { Place, PlaceCandidate } from '../../domain/places';
import type { PlacesGateway } from '../../usecase/placesUseCases';

const toPlace = (d: PlaceDTO): Place => ({
  id: d.id,
  name: d.name,
  label: d.label,
  lat: d.lat,
  lon: d.lon,
  usedBy: d.usedBy.map((u) => ({ ...u })),
});

/** PlacesGateway backed by the REST API (#591). */
export class PlacesApi implements PlacesGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<Place[]> {
    return (await this.http.get<PlaceDTO[]>('/api/places')).map(toPlace);
  }

  async add(p: PlaceCandidate): Promise<Place> {
    const body: PlaceInputDTO = { ...p };
    return toPlace(await this.http.post<PlaceDTO>('/api/places', body));
  }

  async update(id: string, p: PlaceCandidate): Promise<Place> {
    const body: PlaceInputDTO = { ...p };
    return toPlace(await this.http.put<PlaceDTO>(`/api/places/${encodeURIComponent(id)}`, body));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(`/api/places/${encodeURIComponent(id)}`);
  }
}
