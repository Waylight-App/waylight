import type { ElectricityPrices, Zone } from '../domain/electricity';

/** Output port: the electricity price API. */
export interface ElectricityGateway {
  prices(zone: Zone): Promise<ElectricityPrices>;
}

/** Electricity prices for the widget (#288). */
export class ElectricityUseCases {
  constructor(private readonly gateway: ElectricityGateway) {}

  prices(zone: Zone): Promise<ElectricityPrices> {
    return this.gateway.prices(zone);
  }
}
