// Imported from the files, not the feature index, which also exports UI.
import { CompactApi } from '../../features/compact/adapter/api/compactApi';
import { CompactUseCases } from '../../features/compact/usecase/compactUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    compact: CompactUseCases;
  }
}

export const service = defineService('compact', (http) => new CompactUseCases(new CompactApi(http)));
