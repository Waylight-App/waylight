import { BirthdaysApi } from '../../features/birthdays/adapter/api/birthdaysApi';
import { BirthdaysUseCases } from '../../features/birthdays/usecase/birthdaysUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    birthdays: BirthdaysUseCases;
  }
}

export const service = defineService('birthdays', (http) => new BirthdaysUseCases(new BirthdaysApi(http)));
