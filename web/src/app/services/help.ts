// Imported from the files, not the feature index, like the other services.
import { HelpApi } from '../../features/help/adapter/api/helpApi';
import { HelpUseCases } from '../../features/help/usecase/helpUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    help: HelpUseCases;
  }
}

export const service = defineService('help', (http) => new HelpUseCases(new HelpApi(http)));
