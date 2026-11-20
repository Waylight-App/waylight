import { GrafanaApi } from '../../features/grafana/adapter/api/grafanaApi';
import { GrafanaUseCases } from '../../features/grafana/usecase/grafanaUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    grafana: GrafanaUseCases;
  }
}

// Direct imports keep the settings UI out of the initial bundle (#230).
export const service = defineService('grafana', (http) => new GrafanaUseCases(new GrafanaApi(http)));
