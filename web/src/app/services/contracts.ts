import { ContractsApi } from '../../features/contracts/adapter/api/contractsApi';
import { ContractsUseCases } from '../../features/contracts/usecase/contractsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    contracts: ContractsUseCases;
  }
}

export const service = defineService('contracts', (http) => new ContractsUseCases(new ContractsApi(http)));
