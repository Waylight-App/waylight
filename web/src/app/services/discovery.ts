import { DiscoveryApi } from '../../features/discovery/adapter/api/discoveryApi';
import { DiscoveryUseCases } from '../../features/discovery/usecase/discoveryUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    discovery: DiscoveryUseCases;
  }
}

// Services found in the home network for the widget setup (#590).
export const service = defineService('discovery', (http) => new DiscoveryUseCases(new DiscoveryApi(http)));
