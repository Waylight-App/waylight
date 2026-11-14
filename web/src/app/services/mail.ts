import { MailApi } from '../../features/mail/adapter/api/mailApi';
import { MailUseCases } from '../../features/mail/usecase/mailUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    mail: MailUseCases;
  }
}

export const service = defineService('mail', (http) => new MailUseCases(new MailApi(http)));
