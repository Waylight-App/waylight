import type { HttpClient } from '../../../../shared/api-client/http';
import type { PricesDTO } from '../../../../shared/api-client/generated/electricity';
import type { ElectricityPrices, Zone } from '../../domain/electricity';
import type { ElectricityGateway } from '../../usecase/electricityUseCases';

/** ElectricityGateway backed by the REST API. */
export class ElectricityApi implements ElectricityGateway {
  constructor(private readonly http: HttpClient) {}

  async prices(zone: Zone): Promise<ElectricityPrices> {
    const d = await this.http.get<PricesDTO>(`/api/electricity/prices?zone=${encodeURIComponent(zone)}`);
    return {
      zone: d.zone === 'AT' ? 'AT' : 'DE-LU',
      hours: d.hours.map((h) => ({
        start: new Date(h.start),
        centPerKWh: h.centPerKWh,
        cheapest: h.cheapest,
      })),
      fetchedAt: new Date(d.fetchedAt),
      stale: d.stale,
    };
  }
}
