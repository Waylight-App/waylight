import { WikiApi } from '../../features/wiki/adapter/api/wikiApi';
import { WikiUseCases } from '../../features/wiki/usecase/wikiUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    wiki: WikiUseCases;
  }
}

// Direct imports keep UI exports of the feature index out of the initial bundle (#230).
export const service = defineService('wiki', (http) => new WikiUseCases(new WikiApi(http)));
