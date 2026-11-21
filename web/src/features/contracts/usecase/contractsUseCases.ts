import type { ShareGroup, Visibility } from '../../../shared/sharing/sharing';
import { type Contract, type ContractInput, type Period, parseCents } from '../domain/contracts';

/** The fields the backend stores. */
export interface ContractData {
  name: string;
  provider: string;
  costCents: number;
  period: Period;
  termEnd: string;
  noticeDays: number;
  renewMonths: number;
  remindDays: number;
  note: string;
}

/** Output port: contracts backend. */
export interface ContractsGateway {
  list(): Promise<Contract[]>;
  create(data: ContractData, visibility?: Visibility): Promise<Contract>;
  update(id: string, data: ContractData, visibility?: Visibility): Promise<Contract>;
  remove(id: string): Promise<void>;
  shareGroups(): Promise<ShareGroup[]>;
}

const toData = (i: ContractInput): ContractData => ({
  name: i.name.trim(),
  provider: i.provider.trim(),
  costCents: parseCents(i.cost) ?? 0,
  period: i.period,
  termEnd: i.termEnd,
  noticeDays: i.noticeDays,
  renewMonths: i.renewMonths,
  remindDays: i.remindDays,
  note: i.note.trim(),
});

/** Application logic of the contracts (#687). */
export class ContractsUseCases {
  constructor(private readonly gateway: ContractsGateway) {}

  list(): Promise<Contract[]> {
    return this.gateway.list();
  }

  /** Enters (no id) or changes a contract. */
  save(id: string | null, input: ContractInput, visibility?: Visibility): Promise<Contract> {
    const data = toData(input);
    return id ? this.gateway.update(id, data, visibility) : this.gateway.create(data, visibility);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return this.gateway.shareGroups();
  }
}
