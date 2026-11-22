import type { HttpClient } from '../../../../shared/api-client/http';
import type { ContractDTO } from '../../../../shared/api-client/generated/contracts';
import type { ContractInputDTO } from '../../../../shared/api-client/generated/contracts-requests';
import { fetchShareGroups, type ShareGroup, type Visibility } from '../../../../shared/sharing/sharing';
import { type Contract, PERIODS, type Period } from '../../domain/contracts';
import type { ContractData, ContractsGateway } from '../../usecase/contractsUseCases';

const toPeriod = (p: string): Period =>
  (PERIODS as readonly string[]).includes(p) ? (p as Period) : 'monthly';

const toContract = (d: ContractDTO): Contract => ({ ...d, period: toPeriod(d.period) });
const toBody = (d: ContractData, visibility?: Visibility): ContractInputDTO => ({ ...d, visibility });
const path = (id: string) => `/api/contracts/${encodeURIComponent(id)}`;

/** ContractsGateway backed by the REST API. */
export class ContractsApi implements ContractsGateway {
  constructor(private readonly http: HttpClient) {}

  async list(): Promise<Contract[]> {
    return (await this.http.get<ContractDTO[]>('/api/contracts')).map(toContract);
  }

  async create(data: ContractData, visibility?: Visibility): Promise<Contract> {
    return toContract(await this.http.post<ContractDTO>('/api/contracts', toBody(data, visibility)));
  }

  async update(id: string, data: ContractData, visibility?: Visibility): Promise<Contract> {
    return toContract(await this.http.put<ContractDTO>(path(id), toBody(data, visibility)));
  }

  remove(id: string): Promise<void> {
    return this.http.delete(path(id));
  }

  shareGroups(): Promise<ShareGroup[]> {
    return fetchShareGroups(this.http);
  }
}
