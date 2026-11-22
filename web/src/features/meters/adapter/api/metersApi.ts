import type { HttpClient } from '../../../../shared/api-client/http';
import type { DetailDTO, MeterDTO } from '../../../../shared/api-client/generated/meters';
import type { MeterInputDTO } from '../../../../shared/api-client/generated/meters-requests';
import { fetchShareGroups, type ShareGroup, type Visibility } from '../../../../shared/sharing/sharing';
import { type Meter, type MeterDetail, type MeterInput, parseKind, type Reading } from '../../domain/meters';
import type { MetersGateway } from '../../usecase/metersUseCases';

const toMeter = (m: MeterDTO): Meter => ({ ...m, kind: parseKind(m.kind), last: m.last ?? null });
const toBody = (d: MeterInput, visibility?: Visibility): MeterInputDTO => ({ ...d, visibility });
const path = (id: string) => `/api/meters/${encodeURIComponent(id)}`;

/** MetersGateway backed by the REST API. */
export class MetersApi implements MetersGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<Meter[]> {
    return (await this.http.get<MeterDTO[]>('/api/meters')).map(toMeter);
  }

  async detail(id: string, year: number): Promise<MeterDetail> {
    const d = await this.http.get<DetailDTO>(`${path(id)}?year=${year}`);
    return {
      meter: toMeter(d.meter),
      readings: d.readings,
      consumption: {
        ...d.consumption,
        months: d.consumption.months.map((v) => v ?? null),
        total: d.consumption.total ?? null,
      },
      years: d.years,
    };
  }

  async create(data: MeterInput, visibility?: Visibility): Promise<Meter> {
    return toMeter(await this.http.post<MeterDTO>('/api/meters', toBody(data, visibility)));
  }

  async update(id: string, data: MeterInput, visibility?: Visibility): Promise<Meter> {
    return toMeter(await this.http.put<MeterDTO>(path(id), toBody(data, visibility)));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }

  addReading(id: string, reading: Reading): Promise<Reading> {
    return this.http.post<Reading>(`${path(id)}/readings`, reading);
  }

  removeReading(id: string, day: string): Promise<void> {
    return this.http.delete(`${path(id)}/readings/${encodeURIComponent(day)}`);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return fetchShareGroups(this.http);
  }
}
