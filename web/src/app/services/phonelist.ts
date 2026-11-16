import { PhonelistApi } from '../../features/phonelist/adapter/api/phonelistApi';
import { PhonelistUseCases } from '../../features/phonelist/usecase/phonelistUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    phonelist: PhonelistUseCases;
  }
}

// Direct imports keep UI exports of the feature index out of the initial bundle (#230).
export const service = defineService('phonelist', (http) => new PhonelistUseCases(new PhonelistApi(http)));
