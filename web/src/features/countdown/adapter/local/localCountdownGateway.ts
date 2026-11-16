import { newId } from '../../../../shared/id/uuid';
import type { DocumentStore } from '../../../../shared/local/documentStore';
import { MAX_GRACE_DAYS, target, type Countdown, type CountdownInput } from '../../domain/countdown';
import type { CountdownGateway } from '../../usecase/countdownUseCases';

/** A countdown as stored, in the shape of the server's countdown_items table. */
export interface CountdownRow extends Countdown {
  createdAt: string;
}

export interface CountdownDoc {
  items: CountdownRow[];
}

/** When a countdown stops running: its moment, or the end of its day when it has no time. */
function endOf(c: Countdown): Date {
  const end = target(c);
  if (c.time === '') end.setDate(end.getDate() + 1);
  return end;
}

/**
 * Whether a past countdown has outlived graceDays full days after it ended.
 * Mirrors Countdown.Hidden in internal/modules/countdown/domain.
 */
export function hiddenAt(c: Countdown, now: Date, graceDays: number): boolean {
  const end = endOf(c);
  if (now < end) return false;
  const until = new Date(end);
  until.setDate(until.getDate() + graceDays);
  return now >= until;
}

const view = ({ id, title, emoji, date, time }: CountdownRow): Countdown => ({
  id,
  title,
  emoji,
  date,
  time,
});

/**
 * Countdown gateway of the browser extension's local mode (ADR 0019): the
 * server's list rules (grace period, order by moment, then title) on a
 * document of the DocumentStore.
 */
export class LocalCountdownGateway implements CountdownGateway {
  constructor(
    private readonly store: DocumentStore,
    private readonly now: () => Date = () => new Date(),
  ) {}

  private async read(): Promise<CountdownDoc> {
    return (await this.store.get<CountdownDoc>('countdown')) ?? { items: [] };
  }

  async list(graceDays: number): Promise<Countdown[]> {
    const now = this.now();
    const grace = Math.min(Math.max(graceDays, 0), MAX_GRACE_DAYS);
    return (await this.read()).items
      .filter((c) => !hiddenAt(c, now, grace))
      .sort(
        (a, b) =>
          target(a).getTime() - target(b).getTime() || (a.title < b.title ? -1 : a.title > b.title ? 1 : 0),
      )
      .map(view);
  }

  async create(input: CountdownInput): Promise<Countdown> {
    const doc = await this.read();
    const row: CountdownRow = { id: newId(), ...input, createdAt: this.now().toISOString() };
    doc.items.push(row);
    await this.store.put('countdown', doc);
    return view(row);
  }

  async update(id: string, input: CountdownInput): Promise<Countdown> {
    const doc = await this.read();
    const row = doc.items.find((c) => c.id === id);
    if (!row) throw new Error(`countdown ${id} not found`);
    Object.assign(row, input);
    await this.store.put('countdown', doc);
    return view(row);
  }

  async remove(id: string): Promise<void> {
    const doc = await this.read();
    doc.items = doc.items.filter((c) => c.id !== id);
    await this.store.put('countdown', doc);
  }
}
