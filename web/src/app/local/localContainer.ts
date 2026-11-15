/**
 * Composition root of the browser extension's local mode (ADR 0019): the
 * same container as the web app, with local implementations where the
 * extension works without a server. Features that have no local
 * implementation keep their API adapters but are hidden through the module
 * list, so their code is never asked to reach a server.
 */
import { AuthUseCases } from '../../features/auth/usecase/authUseCases';
import { LocalAuthGateway, noCsrf } from '../../features/auth/adapter/local/localAuthGateway';
import { createContainer, type Container } from '../container';
import { localDataKinds, localModules, localWidgetKinds } from './modules';

/** Builds the container for the extension's local mode. */
export function createLocalContainer(): Container {
  // Nothing can expire without a server; an unauthorised answer would mean
  // a hidden feature reached for the network, which the module list prevents.
  const base = createContainer(() => undefined);
  return {
    ...base,
    auth: new AuthUseCases(new LocalAuthGateway(), noCsrf),
    enabledModules() {
      return Promise.resolve({
        modules: new Set(localModules),
        widgetKinds: new Set(localWidgetKinds),
        dataKinds: { ...localDataKinds },
      });
    },
  };
}
