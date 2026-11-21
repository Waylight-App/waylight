import type { FoundService } from '../../../shared/setup/setup';

/** Output port: service discovery of the server (#590). */
export interface DiscoveryGateway {
  find(integration: string): Promise<FoundService[]>;
}

/** Finds services of an integration in the home network for its setup. */
export class DiscoveryUseCases {
  constructor(private readonly gateway: DiscoveryGateway) {}

  /**
   * The services found; an empty list when nothing was found or the viewer
   * may not search (only admins may), so the setup simply asks for the
   * address.
   */
  async find(integration: string): Promise<FoundService[]> {
    try {
      return await this.gateway.find(integration);
    } catch {
      return [];
    }
  }
}
