import { ElectricityApi } from '../../features/electricity/adapter/api/electricityApi';
import { ElectricityUseCases } from '../../features/electricity/usecase/electricityUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    electricity: ElectricityUseCases;
  }
}

export const service = defineService(
  'electricity',
  (http) => new ElectricityUseCases(new ElectricityApi(http)),
);
