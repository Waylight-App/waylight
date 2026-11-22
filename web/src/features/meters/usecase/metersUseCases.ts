import type { ShareGroup, Visibility } from '../../../shared/sharing/sharing';
import type { Meter, MeterDetail, MeterInput, Reading, ReadingInput } from '../domain/meters';

/** Output port: meters backend. */
export interface MetersGateway {
  list(): Promise<Meter[]>;
  detail(id: string, year: number): Promise<MeterDetail>;
  create(data: MeterInput, visibility?: Visibility): Promise<Meter>;
  update(id: string, data: MeterInput, visibility?: Visibility): Promise<Meter>;
  remove(id: string): Promise<void>;
  addReading(id: string, reading: Reading): Promise<Reading>;
  removeReading(id: string, day: string): Promise<void>;
  shareGroups(): Promise<ShareGroup[]>;
}

const toData = (i: MeterInput): MeterInput => ({ name: i.name.trim(), kind: i.kind, unit: i.unit.trim() });

/** Application logic of the meters (#689). */
export class MetersUseCases {
  constructor(private readonly gateway: MetersGateway) {}

  list(): Promise<Meter[]> {
    return this.gateway.list();
  }

  detail(id: string, year: number): Promise<MeterDetail> {
    return this.gateway.detail(id, year);
  }

  /** Creates (no id) or changes a meter. */
  save(id: string | null, input: MeterInput, visibility?: Visibility): Promise<Meter> {
    const data = toData(input);
    return id ? this.gateway.update(id, data, visibility) : this.gateway.create(data, visibility);
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }

  /** Stores a reading; the form has checked that the value is a number. */
  addReading(id: string, input: ReadingInput): Promise<Reading> {
    return this.gateway.addReading(id, { day: input.day, value: input.value ?? 0 });
  }

  removeReading(id: string, day: string): Promise<void> {
    return this.gateway.removeReading(id, day);
  }

  shareGroups(): Promise<ShareGroup[]> {
    return this.gateway.shareGroups();
  }
}
