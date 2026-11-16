import { newId } from '../../../../shared/id/uuid';
import type { DocumentStore } from '../../../../shared/local/documentStore';
import type { FocusDay } from '../../domain/focus';
import type { FocusGateway } from '../../usecase/focusUseCases';

/** Days in the statistics, today included (WeekDays on the server). */
const WEEK_DAYS = 7;

/** A finished session as stored, in the shape of the server's focus_sessions table. */
export interface SessionRow {
  id: string;
  minutes: number;
  /** ISO timestamp. */
  endedAt: string;
}

export interface FocusDoc {
  sessions: SessionRow[];
}

const isoDay = (d: Date): string =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;

/**
 * Sessions and minutes per day for the last seven days including today,
 * oldest first, in the browser's time zone. Mirrors Week in
 * internal/modules/focus/domain.
 */
export function week(sessions: readonly SessionRow[], now: Date): FocusDay[] {
  const days: FocusDay[] = [];
  for (let i = WEEK_DAYS - 1; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() - i);
    days.push({ date: isoDay(d), sessions: 0, minutes: 0 });
  }
  for (const s of sessions) {
    const day = days.find((d) => d.date === isoDay(new Date(s.endedAt)));
    if (day) {
      day.sessions++;
      day.minutes += s.minutes;
    }
  }
  return days;
}

/**
 * Focus gateway of the browser extension's local mode (ADR 0019). The timer
 * runs in the browser anyway; this keeps the finished sessions. Sessions
 * older than the statistics are dropped, as nothing reads them.
 */
export class LocalFocusGateway implements FocusGateway {
  constructor(
    private readonly store: DocumentStore,
    private readonly now: () => Date = () => new Date(),
  ) {}

  private async read(): Promise<FocusDoc> {
    return (await this.store.get<FocusDoc>('focus')) ?? { sessions: [] };
  }

  async stats(): Promise<FocusDay[]> {
    return week((await this.read()).sessions, this.now());
  }

  async record(minutes: number, endedAt: Date): Promise<FocusDay[]> {
    const doc = await this.read();
    const now = this.now();
    const oldest = new Date(now.getFullYear(), now.getMonth(), now.getDate() - (WEEK_DAYS - 1));
    doc.sessions = doc.sessions.filter((s) => new Date(s.endedAt) >= oldest);
    doc.sessions.push({ id: newId(), minutes, endedAt: endedAt.toISOString() });
    await this.store.put('focus', doc);
    return week(doc.sessions, now);
  }
}
