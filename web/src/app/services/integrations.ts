import { ConnectionsApi } from '../../features/integrations/adapter/api/connectionsApi';
import { ConnectionsUseCases } from '../../features/integrations/usecase/connectionsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    integrations: ConnectionsUseCases;
  }
}

export const service = defineService(
  'integrations',
  (http) => new ConnectionsUseCases(new ConnectionsApi(http)),
);
