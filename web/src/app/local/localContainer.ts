/**
 * Composition root of the browser extension's local mode (ADR 0019): the
 * same container as the web app, with local implementations where the
 * extension works without a server. Features that have no local
 * implementation keep their API adapters but are hidden through the module
 * list, so their code is never asked to reach a server.
 */
import { AppearanceUseCases } from '../../features/appearance/usecase/appearanceUseCases';
import { LocalAppearanceGateway } from '../../features/appearance/adapter/local/localAppearanceGateway';
import { AuthUseCases } from '../../features/auth/usecase/authUseCases';
import { LocalAuthGateway, noCsrf } from '../../features/auth/adapter/local/localAuthGateway';
import { DashboardUseCases } from '../../features/dashboard/usecase/dashboardUseCases';
import { LocalDashboardGateway } from '../../features/dashboard/adapter/local/localDashboardGateway';
import { IndexedDbDocumentStore, type DocumentStore } from '../../shared/local/documentStore';
import { createContainer, type Container } from '../container';
import { localDataKinds, localModules, localWidgetKinds } from './modules';

/**
 * Builds the container for the extension's local mode. The store is a
 * parameter so tests can pass one in memory.
 */
export function createLocalContainer(store: DocumentStore = new IndexedDbDocumentStore()): Container {
  // Nothing can expire without a server; an unauthorised answer would mean
  // a hidden feature reached for the network, which the module list prevents.
  const base = createContainer(() => undefined);
  return {
    ...base,
    localMode: true,
    auth: new AuthUseCases(new LocalAuthGateway(), noCsrf),
    dashboard: new DashboardUseCases(new LocalDashboardGateway(store)),
    appearance: new AppearanceUseCases(new LocalAppearanceGateway(store)),
    enabledModules() {
      return Promise.resolve({
        modules: new Set(localModules),
        widgetKinds: new Set(localWidgetKinds),
        dataKinds: { ...localDataKinds },
      });
    },
  };
}
