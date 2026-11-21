// Imported from the files, not the feature index, which also exports UI.
import { AttentionApi } from '../../features/attention/adapter/api/attentionApi';
import { AttentionUseCases } from '../../features/attention/usecase/attentionUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    attention: AttentionUseCases;
  }
}

export const service = defineService('attention', (http) => new AttentionUseCases(new AttentionApi(http)));
